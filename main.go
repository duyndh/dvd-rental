package main

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/internal/config"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func getConf(name string, services []config.Service) *config.Service {
	if len(services) == 0 {
		panic("service is not configured")
	}
	for _, service := range services {
		if service.Name == name {
			return &service
		}
	}
	panic("service does not configured")
}

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	//Get app config
	cfg, err := config.Load("dev")
	if err != nil {
		panic(err)
	}

	svcName := os.Getenv("SERVICE_NAME")
	if len(svcName) == 0 {
		panic("missing SERVICE_NAME")
	}
	//Get Customer configuration
	srvCfg := getConf("customer", cfg.Services)
	cacheCli := redis.NewClient(&redis.Options{
		Addr:     srvCfg.Cache.Addr,
		Password: srvCfg.Cache.Password,
		DB:       0,
	})
	defer cacheCli.Close()
	cacheRepo := customer.NewCacheClient(cacheCli)
	pgConnectionString, err := pg.ParseURL(srvCfg.Database.PSN)
	if err != nil {
		panic(err)
	}
	db := pg.Connect(pgConnectionString)
	defer db.Close()
	
	fielKeys := []string{"method"}
	
	customerRepo := customer.NewCustomerRepository(db, cacheRepo)
	var cs customer.Service
	cs = customer.NewService(customerRepo)
	cs = customer.NewLoggingSerivce(log.With(logger, "service", "customer"), cs)
	cs = customer.NewInstrumentService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name: "request_count",
			Help: "Number of requests received",
		}, fielKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name: "request_latency_microseconds",
			Help: "Request duration",
		}, fielKeys),
		cs,
	)
	
	fmt.Printf("%+v", cfg.Services[0].Cache)

	// customerCache := customer.NewCacheClient()
	var (
	// customer = customer.NewCustomerRepository()
	)

}

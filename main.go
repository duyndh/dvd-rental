package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/internal/config"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err         error
			hostPort    = "localhost:80"
			serviceName = "customer"
			reporter    = zipkinhttp.NewReporter(":9411")
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
	}

	var tracer stdopentracing.Tracer
	{
		logger.Log("tracer", "Zipkin", "type", "Opentracing", "URL", ":9411")
		tracer = zipkinot.Wrap(zipkinTracer)
		zipkinTracer = nil
	}

	//Get app config
	cfg, err := config.Load("dev")
	if err != nil {
		panic(err)
	}

	//Get Customer configuration
	customerCfg := getConf("customer", cfg.Services)
	cacheCli := redis.NewClient(&redis.Options{
		Addr:     customerCfg.Cache.Addr,
		Password: customerCfg.Cache.Password,
		DB:       0,
	})
	defer cacheCli.Close()
	cacheRepo := customer.NewCacheClient(cacheCli)
	pgConnectionString, err := pg.ParseURL(customerCfg.Database.PSN)
	if err != nil {
		panic(err)
	}
	db := pg.Connect(pgConnectionString)
	defer db.Close()

	fielKeys := []string{"method"}
	var historgram metrics.Histogram
	{
		historgram = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name:      "request_latency_microseconds",
			Help:      "Request duration",
		}, fielKeys)
	}
	var counter metrics.Counter
	{
		counter = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name:      "request_count",
			Help:      "Number of requests received",
		}, fielKeys)
	}
	customerRepo := customer.NewCustomerRepository(db, cacheRepo)
	var cs customer.Service
	cs = customer.NewService(customerRepo,logger, counter, historgram)
	customerEndpoint := customer.NewCustomerEndpoint(cs, logger, counter, historgram, tracer)
	// customerHandler = customer.NewHTTPHadne
	mux := http.NewServeMux()

	mux.Handle("/customer/v1/", customer.MakeHandler(customerEndpoint, logger, tracer))

	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", "localhost:9999", "msg", "listening")
		errs <- http.ListenAndServe("localhost:9999", nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("shutting down", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

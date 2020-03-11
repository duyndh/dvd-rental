package main

import (
	"flag"
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
	"github.com/ngray1747/dvd-rental/customer/cache"
	"github.com/ngray1747/dvd-rental/customer/database"
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
	fs := flag.NewFlagSet("dvd_rental", flag.ExitOnError)
	var (
		httpAddr      = fs.String("httpAddr", ":9999", "Http server address")
		zipkinAddr    = fs.String("zipkinAddr", "", "Zipkin tracer address")
		dbUserName    = fs.String("dbUserName", "my-user", "Postgresql username")
		dbPassword    = fs.String("dbPassword", "password123", "Postgresql password")
		dbAddr        = fs.String("dbHost", "", "Postgresql host")
		redisAddr     = fs.String("redisAddr", "", "Redis cache address")
		redisPassword = fs.String("redisPassword", "", "Redis cache password")
	)
	fs.Parse(os.Args[1:])

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	if *zipkinAddr == "" {
		panic("Zipkin configuration required")
	}
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err         error
			hostPort    = "localhost:80"
			serviceName = "customer"
			zipkinURL   = fmt.Sprintf("http://%s/api/v2/spans", *zipkinAddr)
			reporter    = zipkinhttp.NewReporter(zipkinURL)
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
	if *redisAddr == "" {
		panic("Redis configuration required")
	}
	customerCfg := getConf("customer", cfg.Services)
	cacheCli := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       0,
	})
	defer cacheCli.Close()
	if *dbAddr == "" || *dbUserName == "" || *dbPassword == "" {
		panic("Database configuration required")
	}
	cacheRepo := cache.NewCacheClient(cacheCli)
	db := pg.Connect(&pg.Options{
		Addr:     *dbAddr,
		User:     *dbUserName,
		Password: *dbPassword,
		Database: customerCfg.Database.DBName,
	})
	defer db.Close()

	var historgram metrics.Histogram
	{
		historgram = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name:      "request_latency_microseconds",
			Help:      "Request duration",
		}, []string{"method", "success"})
	}
	var counter metrics.Counter
	{
		counter = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "customer_service",
			Name:      "request_count",
			Help:      "Number of requests received",
		}, []string{"method"})
	}
	customerRepo := database.NewCustomerRepository(customerCfg.Cache, db, cacheRepo)
	var cs customer.Service
	cs = customer.NewService(customerRepo, logger, counter, historgram)
	customerEndpoint := customer.NewCustomerEndpoint(cs, tracer)
	// customerHandler = customer.NewHTTPHadne
	mux := http.NewServeMux()

	mux.Handle("/customer/v1/", customer.MakeHandler(customerEndpoint, logger, tracer))

	http.Handle("/", accessControl(mux))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", ":9999", "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, nil)
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

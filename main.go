package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	customerCache "github.com/ngray1747/dvd-rental/customer/cache"
	customerRepo "github.com/ngray1747/dvd-rental/customer/repository"
	"github.com/ngray1747/dvd-rental/dvd"
	dvdCache "github.com/ngray1747/dvd-rental/dvd/cache"
	dvdPB "github.com/ngray1747/dvd-rental/dvd/pb"
	dvdRepo "github.com/ngray1747/dvd-rental/dvd/repository"
	"github.com/ngray1747/dvd-rental/internal/config"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	errServiceNotConfigurated = errors.New("service is not configured")
	errServiceNotFound        = errors.New("service not found")
)

func getConf(name string, services []config.Service) (*config.Service, error) {
	if len(services) == 0 {
		return nil, errServiceNotConfigurated
	}
	for _, service := range services {
		if service.Name == name {
			return &service, nil
		}
	}
	return nil, errServiceNotFound
}

func main() {
	fs := flag.NewFlagSet("dvd_rental", flag.ExitOnError)
	var (
		httpAddr      = fs.String("httpAddr", ":9999", "Http server address")
		grpcAddr      = fs.String("grpcAddr", ":8888", "GRPC server address")
		zipkinAddr    = fs.String("zipkinAddr", "", "Zipkin tracer address")
		dbUserName    = fs.String("dbUserName", "my-user", "Postgresql username")
		dbPassword    = fs.String("dbPassword", "password123", "Postgresql password")
		dbAddr        = fs.String("dbHost", "", "Postgresql host")
		redisAddr     = fs.String("redisAddr", "", "Redis cache address")
		redisPassword = fs.String("redisPassword", "", "Redis cache password")
		svc           = fs.String("service", "", "Service name")
		namespace     = fs.String("namespace", "", "Service namespace")
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
	cacheCli := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       0,
	})
	defer cacheCli.Close()

	if *dbAddr == "" || *dbUserName == "" || *dbPassword == "" {
		panic("Database configuration required")
	}

	var historgram metrics.Histogram
	{
		historgram = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: *namespace,
			Subsystem: *svc,
			Name:      "request_latency_microseconds",
			Help:      "Request duration",
		}, []string{"method", "success"})
	}
	var counter metrics.Counter
	{
		counter = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: *namespace,
			Subsystem: *svc,
			Name:      "request_count",
			Help:      "Number of requests received",
		}, []string{"method"})
	}

	http.Handle("/metrics", promhttp.Handler())
	var grpcServer *grpc.Server
	switch *svc {
	case "customer":
		svcCfg, err := getConf("customer", cfg.Services)
		if err != nil {
			logger.Log("get svc config error: ", err)
			os.Exit(1)
		}
		cacheRepo := customerCache.NewCacheClient(cacheCli)

		db, err := initDB(*dbAddr, *dbUserName, *dbPassword, svcCfg.Database.DBName, []interface{}{customer.Customer{}})
		if err != nil {
			logger.Log("init Db error: ", err)
			os.Exit(1)
		}
		defer db.Close()
		repo := customerRepo.NewCustomerRepository(svcCfg.Cache, db, cacheRepo)
		conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(5*time.Second))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer conn.Close()
		var dvdSvc customer.ProxyService
		dvdSvc = customer.NewProxyMiddleware(conn, context.Background(), tracer, logger)(dvdSvc)
		
		var cs customer.Service
		cs = customer.NewService(repo, logger, counter, historgram, dvdSvc)
		customerEndpoint := customer.NewCustomerEndpoint(cs, tracer)

		mux := http.NewServeMux()
		http.Handle("/", accessControl(mux))
		mux.Handle("/customer/v1/", customer.MakeHandler(customerEndpoint, logger, tracer))
		break
	case "dvd":
		svcCfg, err := getConf("dvd", cfg.Services)
		if err != nil {
			logger.Log("get svc config error: ", err)
			os.Exit(1)
		}
		cacheRepo := dvdCache.NewCacheClient(cacheCli)

		db, err := initDB(*dbAddr, *dbUserName, *dbPassword, svcCfg.Database.DBName, []interface{}{dvd.DVD{}})
		if err != nil {
			logger.Log("init Db error: ", err)
			os.Exit(1)
		}
		defer db.Close()

		repo := dvdRepo.NewDVDRepository(svcCfg.Cache, db, cacheRepo)
		var dvdSrv dvd.Service
		dvdSrv = dvd.NewService(repo, logger, counter, historgram)
		dvdEndpoint := dvd.NewDVDEndpoint(dvdSrv, tracer)
		dvdGRPCServer := dvd.NewGRPCServer(dvdEndpoint, tracer, logger)

		grpcServer = grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
		dvdPB.RegisterDVDRentalServer(grpcServer, dvdGRPCServer)
		// grpcServer.Serve(listener)
		break
	default:
		break
	}

	errs := make(chan error, 2)

	go func() {
		if strings.ToUpper(*namespace) == "API" {
			logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
			errs <- http.ListenAndServe(*httpAddr, nil)
		} else if strings.ToUpper(*namespace) == "SVC" {
			listener, err := net.Listen("tcp", *grpcAddr)
			if err != nil {
				logger.Log("net config error: ", err)
				os.Exit(1)
			}
			logger.Log("transport", "GRPC", "address", *grpcAddr, "msg", "listening")
			errs <- grpcServer.Serve(listener)
		}
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

func initDB(addr, username, password, database string, models []interface{}) (*pg.DB, error) {

	db := pg.Connect(&pg.Options{
		Addr:     addr,
		User:     username,
		Password: password,
		Database: database,
	})

	query := fmt.Sprintf(`SELECT 1 FROM pg_database WHERE datname = '%s';`, database)
	result, err := db.Exec(query)
	if err != nil || result == nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		fmt.Printf("Database \"%s\" is creating...\n", database)
		query := fmt.Sprintf(`CREATE DATABASE %s;`, database)
		_, err := db.Exec(query)
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("Database already existed. Abort migration")

	for _, model := range models {
		fmt.Printf("Creating model:%+v ... \n", model)
		if err := db.CreateTable(model, &orm.CreateTableOptions{
			FKConstraints: true,
			IfNotExists:   true,
		}); err != nil {
			return nil, err
		}
	}

	return db, nil
}

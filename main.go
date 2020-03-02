package main

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/internal/config"
)

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	//Get app config
	cfg, err := config.Load("dev")
	if err != nil {
		panic(err)
	}
	// cli, err := redis.NewClient(&redis.Options{
		
	// })
	// customerCache := customer.NewCacheClient()
	var (
		// customer = customer.NewCustomerRepository()
	)
}
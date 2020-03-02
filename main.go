package main

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/ngray1747/dvd-rental/customer"
)

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestamp)

	
	var (
		customer = customer.NewCustomerRepository()
	)
}
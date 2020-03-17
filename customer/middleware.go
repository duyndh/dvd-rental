package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

type Middleware func(Service) Service

type loggingService struct {
	logger log.Logger
	Service
}

//NewLoggingService init a logging service
func NewLoggingService(logger log.Logger) Middleware {
	return func(svc Service) Service {
		return &loggingService{logger, svc}
	}
}

func (l *loggingService) Register(ctx context.Context, name, address string) (err error) {
	defer func(begin time.Time) {
		l.logger.Log("method", "register", "name", name, "address", address, "error", err, "time", time.Since(begin))
	}(time.Now())
	return l.Service.Register(ctx, name, address)
}

func (l *loggingService) Rent(ctx context.Context, customerID, dvdID string) (err error) {
	defer func(begin time.Time) {
		l.logger.Log("method", "rentDVD", "customerID", customerID, "dvdID", dvdID, "error", err, "time", time.Since(begin))
	}(time.Now())
	return l.Service.Rent(ctx, customerID, dvdID)
}

type instrumentService struct {
	counter   metrics.Counter
	histogram metrics.Histogram
	Service
}

func NewInstrumentService(counter metrics.Counter, histogram metrics.Histogram) Middleware {
	return func(svc Service) Service {
		return &instrumentService{counter: counter, histogram: histogram, Service: svc}
	}
}

func (is *instrumentService) Register(ctx context.Context, name, address string) error {
	err := is.Service.Register(ctx, name, address)
	defer func(begin time.Time) {
		is.counter.With("method", "register").Add(1)
		is.histogram.With("method", "register", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return err
}

func (is *instrumentService) Rent(ctx context.Context, customerID, dvdID string) error {
	err := is.Service.Rent(ctx, customerID, dvdID)
	defer func(begin time.Time) {
		is.counter.With("method", "rentDVD").Add(1)
		is.histogram.With("method", "rentDVD", "success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return err
}
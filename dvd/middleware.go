package dvd

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

type Middleware func(Service) Service

type loggerMiddleware struct {
	logger log.Logger
	svc Service
}

func NewLoggerMiddleware(logger log.Logger) Middleware {
	return func(svc Service) Service {
		return &loggerMiddleware{logger: logger, svc: svc}
	}
}

func (lm *loggerMiddleware) CreateDVD(ctx context.Context, name string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log("method", "CreateDVD", "request_name", name, "error", err, "took", time.Since(begin))
	}(time.Now())
	return lm.svc.CreateDVD(ctx, name)
}

func (lm *loggerMiddleware) RentDVD(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lm.logger.Log("method", "RentDVD", "request_name", id, "error", err, "took", time.Since(begin))
	}(time.Now())
	return lm.svc.RentDVD(ctx, id)
}
type metricMiddleware struct {
	counter metrics.Counter
	histogram metrics.Histogram
	svc Service
}

func NewMetrictMiddleware(counter metrics.Counter, histogram metrics.Histogram) Middleware {
	return func(svc Service) Service {
		return &metricMiddleware{
			counter: counter,
			histogram: histogram,
			svc: svc,
		}
	}
}

func (mw *metricMiddleware) CreateDVD(ctx context.Context, name string) error {
	err := mw.svc.CreateDVD(ctx, name)
	defer func(begin time.Time) {
		mw.counter.With("method", "CreateDVD").Add(1)
		mw.histogram.With("method", "CreateDVD", "success", fmt.Sprint(err!= nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return err
}

func (mw *metricMiddleware) RentDVD(ctx context.Context, id string) error {
	err := mw.svc.RentDVD(ctx, id)
	defer func(begin time.Time) {
		mw.counter.With("method", "RentDVD").Add(1)
		mw.histogram.With("method", "RentDVD", "success", fmt.Sprint(err!= nil)).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return err
}
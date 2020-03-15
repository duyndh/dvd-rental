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
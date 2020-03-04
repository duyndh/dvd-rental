package customer

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)
type Middleware func(Service) Service
type instrumentService struct {
	counter   metrics.Counter
	histogram metrics.Histogram
	Service
}

func NewInstrumentService(counter metrics.Counter, histogram metrics.Histogram) Middleware {
	return func(s Service) Service {
		return &instrumentService{counter: counter, histogram: histogram, Service: s}
	}
}

func (is *instrumentService) Register(ctx context.Context, name, address string) error {
	defer func(begin time.Time) {
		is.counter.With("method", "register").Add(1)
		is.histogram.With("method", "register").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return is.Service.Register(ctx, name, address)
}

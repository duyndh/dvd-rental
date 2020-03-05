package customer

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type registerRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type registerResponse struct {
	Err error `json:"error,omitempty"`
}

func (r registerResponse) error() error { return r.Err }

func makeRegisterEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerRequest)
		err := s.Register(ctx, req.Name, req.Address)
		return registerResponse{Err: err}, nil
	}
}

type CustomerEndpoints struct {
	RegisterEndpoint endpoint.Endpoint
}

//NewCustomerEndpoint wraps all customer service with all middlewares
func NewCustomerEndpoint(cs Service, logger log.Logger, counter metrics.Counter, histogram metrics.Histogram, ot stdopentracing.Tracer) CustomerEndpoints {
	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = makeRegisterEndpoint(cs)
		registerEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(registerEndpoint)
		registerEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(registerEndpoint)
		registerEndpoint = opentracing.TraceServer(ot, "Register")(registerEndpoint)
		registerEndpoint = LoggingMiddleware(log.With(logger, "method", "register"))(registerEndpoint)
		registerEndpoint = InstrumentingMiddleware(counter, histogram)(registerEndpoint)
	}
	return CustomerEndpoints{
		RegisterEndpoint: registerEndpoint,
	}
}

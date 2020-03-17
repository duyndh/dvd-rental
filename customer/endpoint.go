package customer

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
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

type rentRequest struct {
	CustomerID string `json:"customer_id"`
	DVDID      string `json:"dvd_id"`
}

type rentResponse struct {
	Err error `json:"error,omitempty"`
}

func (r rentResponse) error() error { return r.Err }

func makeRentEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(rentRequest)
		err := s.Rent(ctx, req.CustomerID, req.DVDID)
		return rentResponse{Err: err}, nil
	}
}

type CustomerEndpoints struct {
	RegisterEndpoint endpoint.Endpoint
	RentEndpoint endpoint.Endpoint
}

//NewCustomerEndpoint wraps all customer service with all middlewares
func NewCustomerEndpoint(cs Service, ot stdopentracing.Tracer) CustomerEndpoints {
	var registerEndpoint endpoint.Endpoint
	{
		registerEndpoint = makeRegisterEndpoint(cs)
		registerEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(registerEndpoint)
		registerEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(registerEndpoint)
		registerEndpoint = opentracing.TraceServer(ot, "Register")(registerEndpoint)
	}

	var rentEndpoint endpoint.Endpoint
	{
		rentEndpoint = makeRentEndpoint(cs)
		rentEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(rentEndpoint)
		rentEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(rentEndpoint)
		rentEndpoint = opentracing.TraceServer(ot, "Rent")(rentEndpoint)
	}
	
	return CustomerEndpoints{
		RegisterEndpoint: registerEndpoint,
		RentEndpoint: rentEndpoint,
	}
}

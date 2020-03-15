package dvd

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

type CreateDVDRequest struct {
	Name string `json:"name"`
}

type CreateDVDResponse struct {
	Err error `json:"error,omitempty"`
}

func (r CreateDVDResponse) error() error {
	return r.Err
}

func makeCreateDVDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateDVDRequest)
		err := s.CreateDVD(ctx, req.Name)
		return CreateDVDResponse{Err: err}, nil
	}
}

type DVDEndpoints struct {
	CreateDVDEndpoint endpoint.Endpoint
}
func (ep DVDEndpoints) CreateDVD(ctx context.Context, name string) error {
	res, err := ep.CreateDVDEndpoint(ctx, CreateDVDRequest{Name: name})
	if err != nil {
		return err
	}
	response := res.(CreateDVDResponse)
	return response.Err

}
func NewDVDEndpoint(svc Service, ot stdopentracing.Tracer) DVDEndpoints {
	var createDVDEndpoint endpoint.Endpoint
	{
		createDVDEndpoint = makeCreateDVDEndpoint(svc)
		createDVDEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(createDVDEndpoint)
		createDVDEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(createDVDEndpoint)
		createDVDEndpoint = opentracing.TraceServer(ot, "create_dvd")(createDVDEndpoint)
	}
	return DVDEndpoints{
		CreateDVDEndpoint: createDVDEndpoint,
	}
}

package customer

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/ngray1747/dvd-rental/dvd/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type ProxyService interface {
	RentDVD(ctx context.Context, id string) error
}

type proxymw struct {
	context.Context
	Service
	UpdateDVDEndpoint endpoint.Endpoint
}

type rentDVDRequest struct {
	ID string
}

type rentDVDResponse struct {
	Err error
}

func encodeRentDVDRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(rentDVDRequest)
	return &pb.RentDVDRequest{Id: req.ID}, nil
}

func decodeRentDVDResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.RentDVDResponse)
	return rentDVDResponse{Err: strToError(resp.Err)}, nil
}

func strToError(err string) error {
	if err == "" {
		return nil
	}

	return errors.New(err)
}

func NewProxyMiddleware(conn *grpc.ClientConn, ot stdopentracing.Tracer, logger log.Logger) ProxyService {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))

	var opts []grpctransport.ClientOption
	var createDVDEndpoint endpoint.Endpoint
	{
		createDVDEndpoint = grpctransport.NewClient(
			conn,
			"pb.DVDRental",
			"RentDVD",
			encodeRentDVDRequest,
			decodeRentDVDResponse,
			pb.RentDVDResponse{},
			append(opts, grpctransport.ClientBefore(opentracing.ContextToGRPC(ot, logger)))...,
		).Endpoint()
		createDVDEndpoint = opentracing.TraceClient(ot, "CreateDVD")(createDVDEndpoint)
		createDVDEndpoint = limiter(createDVDEndpoint)
		createDVDEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "CreateDVD",
			Timeout: 10 * time.Second,
		}))(createDVDEndpoint)
	}
	
	return func()
}

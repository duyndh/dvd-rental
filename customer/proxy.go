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
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/ngray1747/dvd-rental/dvd/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)
type ProxyMiddleware func(ProxyService) ProxyService
type ProxyService interface {
	UpdateDVDStatus(ctx context.Context, DVDID string) error
}

type proxymw struct {
	context.Context
	ProxyService
	UpdateDVDStatusEndpoint endpoint.Endpoint
}

type updateDVDStatusRequest struct {
	ID string
}

type updateDVDStatusResponse struct {
	Err error
}

func encodeRentDVDRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(updateDVDStatusRequest)
	return &pb.RentDVDRequest{Id: req.ID}, nil
}

func decodeRentDVDResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.RentDVDResponse)
	return updateDVDStatusResponse{Err: strToError(resp.Err)}, nil
}

func strToError(err string) error {
	if err == "" {
		return nil
	}

	return errors.New(err)
}

func (pm proxymw) UpdateDVDStatus(ctx context.Context, DVDID string) error {
	response, err := pm.UpdateDVDStatusEndpoint(ctx, updateDVDStatusRequest{
		ID: DVDID,
	})
	if err != nil {
		return err
	}
	resp := response.(updateDVDStatusResponse)
	return resp.Err
}

func NewProxyMiddleware(conn *grpc.ClientConn, ctx context.Context, ot stdopentracing.Tracer, logger log.Logger) ProxyMiddleware {
	return func(svc ProxyService) ProxyService {
		limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))

		var opts []grpctransport.ClientOption
		var rentDVDEndpoint endpoint.Endpoint
		{
			rentDVDEndpoint = grpctransport.NewClient(
				conn,
				"pb.DVDRental",
				"RentDVD",
				encodeRentDVDRequest,
				decodeRentDVDResponse,
				pb.RentDVDResponse{},
				append(opts, grpctransport.ClientBefore(opentracing.ContextToGRPC(ot, logger)))...,
			).Endpoint()
			rentDVDEndpoint = opentracing.TraceClient(ot, "RentDVD")(rentDVDEndpoint)
			rentDVDEndpoint = limiter(rentDVDEndpoint)
			rentDVDEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
				Name:    "RentDVD",
				Timeout: 10 * time.Second,
			}))(rentDVDEndpoint)
		}
		return proxymw{ctx, svc, rentDVDEndpoint}
	}
}

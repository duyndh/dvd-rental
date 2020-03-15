package dvd

import (
	"context"
	"time"
	"errors"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/ngray1747/dvd-rental/dvd/pb"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type grpcServer struct {
	createDVD grpctransport.Handler
}

func (g *grpcServer) CreateDVD(ctx context.Context, req *pb.CreateDVDRequest) (*pb.CreateDVDResponse, error) {
	_, res, err := g.createDVD.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.(*pb.CreateDVDResponse), nil
}

func decodeGRPCCreateDVDRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.CreateDVDRequest)
	return CreateDVDRequest{Name: req.Name}, nil
}

func encodeGRPCCreateDVDRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(CreateDVDRequest)
	return &pb.CreateDVDRequest{Name: req.Name}, nil
}

func decodeGRPCCreateDVDResponse(_ context.Context, response interface{}) (interface{}, error) {
	res := response.(*pb.CreateDVDResponse)
	return CreateDVDResponse{Err: strToError(res.Err)}, nil
}

func encodeGRPCCreateDVDResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(CreateDVDResponse)
	return &pb.CreateDVDResponse{Err: errToString(resp.Err)}, nil
}

func errToString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
func strToError(err string) error {
	if err == "" {
		return nil
	}

	return errors.New(err)
}
func NewGRPCServer(endpoints DVDEndpoints, ot stdopentracing.Tracer, logger log.Logger) pb.DVDRentalServer {
	opts := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	createDVDHandler := grpctransport.NewServer(
		endpoints.CreateDVDEndpoint,
		decodeGRPCCreateDVDRequest,
		encodeGRPCCreateDVDResponse,
		append(opts, grpctransport.ServerBefore(opentracing.GRPCToContext(ot, "create DVD", logger)))...,
	) 
	return &grpcServer{
		createDVDHandler,
	}
}

func NewGRPCClient(connection *grpc.ClientConn, ot stdopentracing.Tracer, logger log.Logger) Service {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))

	var opts []grpctransport.ClientOption
	var createDVDEndpoint endpoint.Endpoint
	{
		createDVDEndpoint = grpctransport.NewClient(
			connection,
			"pb.DVDRental",
			"CreateDVD",
			encodeGRPCCreateDVDRequest,
			decodeGRPCCreateDVDResponse,
			pb.CreateDVDResponse{},
			append(opts, grpctransport.ClientBefore(opentracing.ContextToGRPC(ot, logger)))...,
		).Endpoint()
		createDVDEndpoint = opentracing.TraceClient(ot,"CreateDVD")(createDVDEndpoint)
		createDVDEndpoint = limiter(createDVDEndpoint)
		createDVDEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name: "CreateDVD",
			Timeout: 10 * time.Second,
		}))(createDVDEndpoint)
	}
	return DVDEndpoints{
		CreateDVDEndpoint: createDVDEndpoint,
	}
}
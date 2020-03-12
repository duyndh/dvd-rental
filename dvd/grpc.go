package dvd

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/ngray1747/dvd-rental/dvd/pb"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
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
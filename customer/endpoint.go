package customer

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type registerRequest struct {
	Name string `json:"name"`
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
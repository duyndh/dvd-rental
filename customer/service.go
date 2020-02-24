package customer

import (
	"context"
	"errors"
)

var (
	errInvalidArgument = errors.New("invalid argument(s)")
)
//Service describe customer business
type Service interface {
	Register(ctx context.Context, name, address string) (string, error)
	Rent(ctx context.Context, id int) error
	Buy(ctx context.Context, id int) error
	Return(ctx context.Context, id int) error
}

//NewService init customer's service interface
func NewService() Service {
	var svc Service
	return svc
}

//customerService implement Service interface
type customerService struct{}

func newCustomerService() customerService {
	return customerService{}
}

func (c *customerService) Register(ctx context.Context, name, address string) (string, error) {
	if name == "" || address == "" {
		return "", errInvalidArgument
	}
	
}

package customer

import (
	"context"
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

var (
	errInvalidArgument = errors.New("invalid argument(s)")
)

//Service describe customer business
type Service interface {
	//Register customer
	Register(ctx context.Context, name, address string) error
	//Customer rent a dvd
	// Rent(ctx context.Context, id int) error
	//Customer buys a dvd
	// Buy(ctx context.Context, id int) error
	//Customer returns borrowed dvd
	// Return(ctx context.Context, id int) error
}

//NewService return customerService with all expected function
func NewService(customerRepo Repository, logger log.Logger, counter metrics.Counter, histogram metrics.Histogram) Service {
	var svc Service
	{
		svc = NewCustomerService(customerRepo)
		svc = NewLoggingService(logger)(svc)
		svc = NewInstrumentService(counter, histogram)(svc)
	}
	return svc
}

//customerService implement Service interface
type customerService struct {
	repo Repository
}

//NewCustomerService init customer's service interface
func NewCustomerService(customerRepo Repository) Service {
	return &customerService{repo: customerRepo}
}

func (c *customerService) Register(ctx context.Context, name, address string) error {
	if name == "" || address == "" {
		return errInvalidArgument
	}
	customer, err := NewCustomer(name, address)
	if err != nil {
		return err
	}

	if err := c.repo.Store(customer); err != nil {
		return err
	}
	return nil

}

// //TODO: Need implement
// func (c *customerService) Rent(ctx context.Context, id int) error {
// 	return nil
// }

// //TODO: Need implement
// func (c *customerService) Buy(ctx context.Context, id int) error {
// 	return nil
// }

// //TODO: Need implement
// func (c *customerService) Return(ctx context.Context, id int) error {
// 	return nil
// }

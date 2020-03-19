package customer

import (
	"github.com/google/uuid"
	"github.com/ngray1747/dvd-rental/internal/model"
)


//Customer represents the customer model
type Customer struct {
	model.Base
	Name      string `pg:",notnull"`
	Address   string `pg:",notnull"`
}

//Repository represent database/cache business
type Repository interface {
	Store(c *Customer) error
	GetByID(q string) (*Customer, error)
	Update(c *Customer) error
	Delete(c *Customer) error
}

//NewCustomer init a new customer with name and address.
func NewCustomer(name, address string) (*Customer, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Customer{
		Base: model.Base{
			ID:      id.String(),
		},
		Name:    name,
		Address: address,
	}, nil
}

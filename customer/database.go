package customer

import (
	"github.com/go-pg/pg/v9"
)

type CustomerRepository interface {
	Store(c *Customer) error
	GetByID(q string) (*Customer, error)
	Update(c *Customer) error
	Delete(c *Customer) error
}

type customerRepository struct {
	db *pg.DB
}
//NewCustomerRepository create a new customer repository.
func NewCustomerRepository(db *pg.DB) CustomerRepository {
	return &customerRepository{ db:db }
}

func (cr *customerRepository)Store(c *Customer) error {
	// tx, err := cr.db.Begin()
	return nil
}


func (cr *customerRepository)GetByID(q string) (*Customer, error) {
	// tx, err := cr.db.Begin()
	return nil,nil
}


func (cr *customerRepository)Update(c *Customer) error {
	// tx, err := cr.db.Begin()
	return nil
}

func (cr *customerRepository)Delete(c *Customer) error {
	// tx, err := cr.db.Begin()
	return nil
}
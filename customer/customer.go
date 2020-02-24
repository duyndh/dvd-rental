package customer

import (
	"time"
	"context"

	"github.com/google/uuid"
	"github.com/go-pg/pg/v9/orm"
)

var _ orm.BeforeInsertHook = (*Base)(nil)
var _ orm.BeforeUpdateHook = (*Base)(nil)

// BeforeInsert hooks into insert operations, setting createdAt and updatedAt to current time
func (c *Customer) BeforeInsert(c context.Context) (context.Context, error) {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	return c, nil
}

// BeforeUpdate hooks into update operations, setting updatedAt to current time
func (c *Customer) BeforeUpdate(c context.Context) (context.Context, error) {
	c.UpdatedAt = time.Now()
	return c, nil
}

type Customer struct {
	ID string `pg:type:uuid`
	Name string `pg:",notnull"`
	Address string `pg:",notnull"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `pg:",soft_delete"`
}

type Repository interface {
	Store(c *Customer) error
	GetByID(q string) (*Customer, error)
	Update(c *Customer) error
	Delete(c *Customer) error
}

func NewCustomer(name, address string) (*Customer, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Customer{
		ID: id,
		Name: name,
		Address: address,
	}
}

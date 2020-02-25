package customer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/go-pg/pg/v9/orm"
)

var _ orm.BeforeInsertHook = (*Customer)(nil)
var _ orm.BeforeUpdateHook = (*Customer)(nil)

// BeforeInsert hooks into insert operations, setting createdAt and updatedAt to current time
func (c *Customer) BeforeInsert(ctx context.Context) (context.Context, error) {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	return ctx, nil
}

// BeforeUpdate hooks into update operations, setting updatedAt to current time
func (c *Customer) BeforeUpdate(ctx context.Context) (context.Context, error) {
	c.UpdatedAt = time.Now()
	return ctx, nil
}

//Customer represents the customer model
type Customer struct {
	ID        string `pg:type:uuid`
	Name      string `pg:",notnull"`
	Address   string `pg:",notnull"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `pg:",soft_delete"`
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
		ID:      id.String(),
		Name:    name,
		Address: address,
	}, nil
}

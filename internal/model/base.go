package model

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9/orm"
)

// Base contains common fields for all tables
type Base struct {
	ID        string       `pg:",pk" json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	DeletedAt time.Time `pg:",soft_delete"json:"deleted_at,omitempty"`
}

var _ orm.BeforeInsertHook = (*Base)(nil)
var _ orm.BeforeUpdateHook = (*Base)(nil)

// BeforeInsert hooks into insert operations, setting createdAt and updatedAt to current time
func (b *Base) BeforeInsert(c context.Context) (context.Context, error) {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return c, nil
}

// BeforeUpdate hooks into update operations, setting updatedAt to current time
func (b *Base) BeforeUpdate(c context.Context) (context.Context, error) {
	b.UpdatedAt = time.Now()
	return c, nil
}

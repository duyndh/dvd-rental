package customer

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
)

type customerRepository struct {
	db *pg.DB
}

type customerCache struct {
	repo Repository
	cache redis.Client
}

// func (cr *customerRepository) Store(c *Customer) error {
// 	return nil
// }
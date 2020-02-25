package customer

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/internal/config"
)

//Cache provides access to customer cache
type Cache interface {
	StoreToCache(key string, value Customer) error
	GetFromCache(key, field string) (*Customer, error)
}
type customerRepository struct {
	cfg   *config.Cache
	db    *pg.DB
	cache Cache
}

//NewCustomerRepository create a new customer repository.
func NewCustomerRepository(db *pg.DB, cache Cache) Repository {
	return &customerRepository{db: db, cache: cache}
}

func (cr *customerRepository) Store(c *Customer) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()
	err = tx.Insert(c)
	if err != nil {
		return err
	}
	err = cr.cache.StoreToCache(cr.cfg.CacheKey, *c)
	if err != nil {
		return err
	}
	return nil
}

func (cr *customerRepository) GetByID(id string) (*Customer, error) {
	//* Get data from cache first
	customer, err := cr.cache.GetFromCache(cr.cfg.CacheKey, id)
	if err != nil && err != redis.Nil {
		return nil, err
	}else if err == redis.Nil {
		customer = &Customer{ID: id}
		if err := cr.db.Select(customer); err != nil {
			return nil, err
		}
	}
	return customer, nil
}

func (cr *customerRepository) Update(c *Customer) error {
	// tx, err := cr.db.Begin()
	return nil
}

func (cr *customerRepository) Delete(c *Customer) error {
	// tx, err := cr.db.Begin()
	return nil
}

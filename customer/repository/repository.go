package repository

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/internal/config"
)

//Cache provides access to customer cache
type Cache interface {
	StoreToCache(key string, value customer.Customer) error
	GetFromCache(key, field string) (*customer.Customer, error)
	RemoveFromCache(key, field string) error
}
type customerRepository struct {
	cfg   *config.Cache
	db    *pg.DB
	cache Cache
}

//NewCustomerRepository create a new customer repository.
func NewCustomerRepository(cfg *config.Cache, db *pg.DB, cache Cache) customer.Repository {
	return &customerRepository{cfg: cfg, db: db, cache: cache}
}

func (cr *customerRepository) Store(c *customer.Customer) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()
	if err := tx.Insert(c); err != nil {
		return err
	}

	if err := cr.cache.StoreToCache(cr.cfg.CacheKey, *c); err != nil {
		return err
	}
	return tx.Commit()
}

func (cr *customerRepository) GetByID(id string) (*customer.Customer, error) {
	//* Get data from cache first
	cus, err := cr.cache.GetFromCache(cr.cfg.CacheKey, id)
	if err != nil && err != redis.Nil {
		return nil, err
	} else if err == redis.Nil {
		cus = &customer.Customer{ID: id}
		// Get from database
		if err := cr.db.Select(cus); err != nil {
			return nil, err
		}
		// Set back to cache
		if err = cr.cache.StoreToCache(cr.cfg.CacheKey, *cus); err != nil {
			return nil, err
		}
	}
	return cus, nil
}

func (cr *customerRepository) Update(c *customer.Customer) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()

	if err := tx.Update(c); err != nil {
		return err
	}

	if err := cr.cache.StoreToCache(cr.cfg.CacheKey, *c); err != nil {
		return err
	}

	return nil
}

func (cr *customerRepository) Delete(c *customer.Customer) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()

	if err := tx.Delete(c); err != nil {
		return err
	}

	if err := cr.cache.RemoveFromCache(cr.cfg.CacheKey, c.ID); err != nil {
		return err
	}

	return nil
}

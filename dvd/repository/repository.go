package repository

import (
	"github.com/go-pg/pg/v9"
	"github.com/ngray1747/dvd-rental/dvd"
	"github.com/ngray1747/dvd-rental/internal/config"
)

type Cache interface {
	StoreToCache(key string, value dvd.DVD) error
}

type dvdRepository struct {
	cfg   *config.Cache
	db    *pg.DB
	cache Cache
}

//NewDVDRepository create a new dvd repository.
func NewDVDRepository(cfg *config.Cache, db *pg.DB, cache Cache) dvd.Repository {
	return &dvdRepository{cfg: cfg, db: db, cache: cache}
}

func (cr *dvdRepository) Store(d *dvd.DVD) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()
	if err := tx.Insert(d); err != nil {
		return err
	}

	if err := cr.cache.StoreToCache(cr.cfg.CacheKey, *d); err != nil {
		return err
	}
	return tx.Commit()
}

package repository

import (
	"errors"

	"github.com/go-pg/pg/v9"
	"github.com/ngray1747/dvd-rental/dvd"
	"github.com/ngray1747/dvd-rental/internal/config"
)

var (
	errDVDNotAvailable = errors.New("dvd not available")
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

func (cr *dvdRepository) Update(id string, status dvd.Status) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Close()
	d := new(dvd.DVD)
	if err := tx.Model(d).Where("id = ?", id).Select(); err != nil {
		return err
	}
	
	if d.Status == dvd.NotAvailable {
		return errDVDNotAvailable
	}

	d.Status = status
	if err := tx.Update(d); err != nil {
		return err
	}
	
	if err := cr.cache.StoreToCache(cr.cfg.CacheKey, *d); err != nil {
		return err
	}
	return tx.Commit()
}

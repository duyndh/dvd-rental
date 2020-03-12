package dvd

import (
	"context"
	"errors"
)

var (
	errInvalidDVDName = errors.New("invalid dvd name")
)
type Service interface {
	CreateDVD(ctx context.Context, name string) error
}


type dvdService struct {
	repo  Repository
}

func NewDVDService(dvdRepo Repository) Service {
	return &dvdService{repo: dvdRepo}
}

func (d *dvdService) CreateDVD(ctx context.Context, name string) error {
	if name == "" {
		return errInvalidDVDName
	}

	dvd, err := NewDVD(name)
	if err != nil {
		return err
	}

	return d.repo.Store(dvd)
}
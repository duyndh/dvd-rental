package dvd

import (
	"github.com/google/uuid"
	"github.com/ngray1747/dvd-rental/internal/model"
)

type Status uint8

const (
	Available = iota + 1
	NotAvailable
)

var Statuss = []Status {
	Available,
	NotAvailable,
}

func (s Status) ToString() string {
	switch s {
	case Available:
		return "Available"
	case NotAvailable:
		return "NotAvailable"
	default:
		return "Unknown"
	}
}

type DVD struct {
	model.Base
	Name string `pg:",notnull"`
	Status Status 
}

type Repository interface {
	Store(dvd *DVD) error
}

//NewDVD generate a dvd model with input name
func NewDVD(name string) (*DVD, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &DVD{
		Base: model.Base{
			ID: id.String(),
		},
		// ID:      id.String(),
		Name:    name,
		Status: Available,
	}, nil
}
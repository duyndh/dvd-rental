package dvd_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/ngray1747/dvd-rental/dvd"
	"github.com/ngray1747/dvd-rental/dvd/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	repo := new(mocks.Repository)
	svc := dvd.NewService(repo, log.NewNopLogger(), discard.NewCounter(), discard.NewHistogram())
	type args struct {
		name string
	}
	cases := []struct {
		name    string
		args    args
		wantErr bool
		mock    func()
	}{
		{
			name: "OK",
			args: args{
				name: "Title 1",
			},
			wantErr: false,
			mock: func() {
				repo.On("Store", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "missing name",
			args: args{
				name: "",
			},
			wantErr: true,
			mock:    func() {},
		},
		{
			name: "store failed",
			args: args{
				name: "Title 2",
			},
			wantErr: true,
			mock: func() {
				repo.On("Store", mock.Anything).Return(errors.New("store failed")).Once()
			},
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			v.mock()
			err := svc.CreateDVD(ctx, v.args.name)
			assert.Equalf(v.wantErr, err != nil, "name: %v , wantErr %v, got %v , err ", v.name, v.wantErr, err != nil, err)
		})
	}
}

func TestRentDVD(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	repo := new(mocks.Repository)
	svc := dvd.NewService(repo, log.NewNopLogger(), discard.NewCounter(), discard.NewHistogram())
	type args struct {
		id string
	}
	cases := []struct {
		name    string
		args    args
		wantErr bool
		mock    func()
	}{
		{
			name: "OK",
			args: args{
				id: "5e8b83c9-36f3-4084-94b5-33153246d534",
			},
			wantErr: false,
			mock: func() {
				repo.On("Update", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "missing id",
			args: args{
				id: "",
			},
			wantErr: true,
			mock:    func() {},
		},
		{
			name: "Update failed",
			args: args{
				id: "Title 2",
			},
			wantErr: true,
			mock: func() {
				repo.On("Update", mock.Anything).Return(errors.New("Update failed")).Once()
			},
		},
		{
			name: "id failed",
			args: args{
				id: "some-id",
			},
			wantErr: true,
			mock:    func() {},
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			v.mock()
			err := svc.RentDVD(ctx, v.args.id)
			assert.Equalf(v.wantErr, err != nil, "name: %v , wantErr %v, got %v , err ", v.name, v.wantErr, err != nil, err)
		})
	}
}

package customer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/customer/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	repo := new(mocks.Repository)
	svc := customer.NewService(repo, log.NewNopLogger(), discard.NewCounter(), discard.NewHistogram())
	type args struct {
		name    string
		address string
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
				name:    "Duynguyen",
				address: "1102 Truong Sa Street",
			},
			wantErr: false,
			mock: func() {
				repo.On("Store", mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "missing name",
			args: args{
				address: "1102 Truong Sa Street",
			},
			wantErr: true,
			mock: func() {},
		},
		{
			name: "missing address",
			args: args{
				name: "Duynguyen",
			},
			wantErr: true,
			mock: func() {},
		},
		{
			name: "store failed",
			args: args{
				name:    "Duynguyen",
				address: "1102 Truong Sa Street",
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
			err := svc.Register(ctx, v.args.name, v.args.address)
			assert.Equalf(v.wantErr, err != nil, "name: %v , wantErr %v, got %v , err ", v.name, v.wantErr, err != nil, err)
		})
	}
}

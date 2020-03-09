package customer_test

import (
	"context"
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
	repo := &mocks.Repository{}
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
				repo.On("Store", mock.Anything).Return(nil)
			},
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			v.mock()
			err := svc.Register(ctx, v.args.name, v.name)
			assert.Equal(v.wantErr, err != nil)
		})
	}
}

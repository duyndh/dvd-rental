package cache_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/customer/cache"
)

var db *redis.Client
func TestMain(m *testing.M) {

	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("bitnami/redis", "latest", []string{"ALLOW_EMPTY_PASSWORD=yes"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		db = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
		})

		return db.Ping().Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	code := m.Run()
	// When you're done, kill and remove the container
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func TestStoreToCache(t *testing.T) {
	client := cache.NewCacheClient(db)
	type args struct {
		key      string
		customer customer.Customer
	}
	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				key: "customers",
				customer: customer.Customer{
					ID:        "66d112da-07e3-41de-bce3-86fe2bd52b24",
					Name:      "Duy Nguyen",
					Address:   "1102 Truong Sa Street",
					CreatedAt: time.Now(),
				},
			},
			wantErr: false,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			err := client.StoreToCache(v.args.key, v.args.customer)
			assert.Equal(t, v.wantErr, err != nil)
		})
	}
}

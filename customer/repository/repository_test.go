package repository_test

import (
	// "database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/customer/cache"
	"github.com/ngray1747/dvd-rental/customer/repository"
	"github.com/ngray1747/dvd-rental/internal/config"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
)

var db *pg.DB
var cacheClient *redis.Client
var cacheConfig *config.Cache

func TestMain(m *testing.M) {

	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("bitnami/postgresql", "latest", []string{"POSTGRESQL_USERNAME=my_user", "POSTGRESQL_PASSWORD=password123", "POSTGRESQL_DATABASE=dvd_rental"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		pgConnectionString, err := pg.ParseURL(fmt.Sprintf("postgres://my_user:password123@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), "dvd_rental"))
		if err != nil {
			panic(err)
		}
		db = pg.Connect(pgConnectionString)
		_, err = db.Exec(`CREATE TABLE public.customers (
			id uuid NOT NULL,
			name varchar(55) NULL,
			address varchar(255) NULL,
			created_at timestamptz NULL,
			updated_at timestamptz NULL,
			deleted_at timestamptz NULL,
			CONSTRAINT customers_pkey PRIMARY KEY (id)
		)`)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	cacheResource, err := pool.Run("bitnami/redis", "latest", []string{"ALLOW_EMPTY_PASSWORD=yes"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		cacheClient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", cacheResource.GetPort("6379/tcp")),
		})

		return cacheClient.Ping().Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	cacheConfig = &config.Cache{
		CacheKey: "customers",
	}

	code := m.Run()

	// When you're done, kill and remove the container
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := pool.Purge(cacheResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func TestStore(t *testing.T) {
	cacheCli := cache.NewCacheClient(cacheClient)
	repo := repository.NewCustomerRepository(cacheConfig, db, cacheCli)
	type args struct {
		customer *customer.Customer
	}
	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				customer: &customer.Customer{
					ID:      "18eb0b6e-8757-4dfb-b062-1c7944e2b8f7",
					Name:    "Duy Nguyen",
					Address: "1102 Truong Sa Street",
				},
			},
			wantErr: false,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			err := repo.Store(v.args.customer)
			assert.Equal(t, v.wantErr, err != nil)
		})
	}
}

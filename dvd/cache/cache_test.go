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
	"github.com/ngray1747/dvd-rental/dvd"
	"github.com/ngray1747/dvd-rental/dvd/cache"
	"github.com/ngray1747/dvd-rental/internal/model"
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
		dvd dvd.DVD
	}
	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				key: "dvds",
				dvd: dvd.DVD{
					Base: model.Base{
						ID:        "66d112da-07e3-41de-bce3-86fe2bd52b24",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Title 1",
				},
			},
			wantErr: false,
		},
	}
	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			err := client.StoreToCache(v.args.key, v.args.dvd)
			assert.Equal(t, v.wantErr, err != nil)
		})
	}
}

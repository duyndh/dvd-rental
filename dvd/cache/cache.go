package cache

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/ngray1747/dvd-rental/dvd"
	"github.com/ngray1747/dvd-rental/dvd/repository"
	"github.com/vmihailenco/msgpack"
)

var errNilResult = errors.New("nil value")

type cacheClient struct {
	client *redis.Client
}

//NewCacheClient init a new cache client
func NewCacheClient(cli *redis.Client) repository.Cache {
	return &cacheClient{client: cli}
}

func (c *cacheClient) StoreToCache(key string, dvd dvd.DVD) error {
	bytes, err := msgpack.Marshal(dvd)
	if err != nil {
		return err
	}

	return c.client.HSet(key, dvd.ID, bytes).Err()
}

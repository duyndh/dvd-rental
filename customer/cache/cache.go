package cache

import (
	"encoding/json"
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/vmihailenco/msgpack"
	"github.com/ngray1747/dvd-rental/customer"
	"github.com/ngray1747/dvd-rental/customer/database"
)

var errNilResult = errors.New("nil value")

type cacheClient struct {
	client *redis.Client
}

//NewCacheClient init a new cache client
func NewCacheClient(cli *redis.Client) database.Cache {
	return &cacheClient{client: cli}
}

func (c *cacheClient) StoreToCache(key string, customer customer.Customer) error {
	bytes, err := msgpack.Marshal(customer)
	if err != nil {
		return err
	}
	
	return  c.client.HSet(key, customer.ID, bytes).Err()
}

func (c *cacheClient) GetFromCache(key, field string) (*customer.Customer, error) {
	val, err := c.client.HGet(key, field).Bytes()
	if err != nil {
		return nil, err
	}

	var result = new(customer.Customer)
	err = json.Unmarshal(val, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *cacheClient) RemoveFromCache(key, field string) error {
	if _, err := c.client.HDel(key, field).Result(); err != nil {
		return err
	}
	return nil
}
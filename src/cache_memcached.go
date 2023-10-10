package gormcache

import (
	"context"
	"encoding/json"
	"time"

	memcache "github.com/bradfitz/gomemcache/memcache"
)

// MemcacheClient is a wrapper for gomemcache client
type MemcacheClient struct {
	client *memcache.Client
}

// NewMemcacheClient returns a new RedisClient instance
func NewMemcacheClient(client *memcache.Client) *MemcacheClient {
	return &MemcacheClient{
		client: client,
	}
}

// Get gets value from memcache by key using json encoding/decoding
func (r *MemcacheClient) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(key)
	if err != nil {
		return nil, nil
	}
	value := data.Value
	//log.Printf("get cache %v, key: %v", value, key)
	return value, nil
}

// Set sets value to memcache by key with ttl using json encoding/decoding
func (r *MemcacheClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value) // encode value to json bytes using json encoding/decoding
	if err != nil {
		return err
	}
	return r.client.Set(&memcache.Item{Key: key, Value: data, Expiration: int32(ttl.Seconds())})
}

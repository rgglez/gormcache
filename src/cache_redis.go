package gormcache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redis "github.com/go-redis/redis/v8"
)

// RedisClient is a wrapper for go-redis client
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient returns a new RedisClient instance
func NewRedisClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
	}
}

// Get gets value from redis by key using json encoding/decoding
func (r *RedisClient) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	//log.Printf("get cache, key: %v", key)
	return data, nil
}

// Set sets value to redis by key with ttl using json encoding/decoding
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value) // encode value to json bytes using json encoding/decoding
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

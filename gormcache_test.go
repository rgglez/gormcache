/*
   Copyright 2023 evangwt
   Copyright 2023 Rodolfo González González for the modifications in
   this fork.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package gormcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormcache "github.com/rgglez/gormcache"
)

// mockCacheClient is an in-memory CacheClient for unit testing.
type mockCacheClient struct {
	store map[string][]byte
	gets  int
	sets  int
}

func newMockCacheClient() *mockCacheClient {
	return &mockCacheClient{store: make(map[string][]byte)}
}

func (m *mockCacheClient) Get(_ context.Context, key string) (interface{}, error) {
	m.gets++
	v, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockCacheClient) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.sets++
	if b, ok := value.([]byte); ok {
		m.store[key] = b
		return nil
	}
	return nil
}

func TestNewGormCache(t *testing.T) {
	client := newMockCacheClient()
	config := gormcache.CacheConfig{TTL: 30 * time.Second, Prefix: "test:"}
	cache := gormcache.NewGormCache("test_cache", client, config)
	assert.NotNil(t, cache)
	assert.Equal(t, "test_cache", cache.Name())
}

func TestCacheConfig(t *testing.T) {
	config := gormcache.CacheConfig{
		TTL:    10 * time.Minute,
		Prefix: "prefix:",
	}
	assert.Equal(t, 10*time.Minute, config.TTL)
	assert.Equal(t, "prefix:", config.Prefix)
}

func TestContextKeys(t *testing.T) {
	ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
	useCache, ok := ctx.Value(gormcache.UseCacheKey).(bool)
	assert.True(t, ok)
	assert.True(t, useCache)

	ctx2 := context.WithValue(ctx, gormcache.CacheTTLKey, 5*time.Second)
	ttl, ok := ctx2.Value(gormcache.CacheTTLKey).(time.Duration)
	assert.True(t, ok)
	assert.Equal(t, 5*time.Second, ttl)
}

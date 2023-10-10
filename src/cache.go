package gormcache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm/callbacks"

	"gorm.io/gorm"
)

var (
	UseCacheKey struct{}
	CacheTTLKey struct{}
)

// CacheClient is an interface for cache operations
type CacheClient interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// CacheConfig is a struct for cache options
type CacheConfig struct {
	TTL    time.Duration // cache expiration time
	Prefix string        // cache key prefix
}

// GormCache is a cache plugin for gorm
type GormCache struct {
	name   string
	client CacheClient
	config CacheConfig
}

// NewGormCache returns a new GormCache instance
func NewGormCache(name string, client CacheClient, config CacheConfig) *GormCache {
	return &GormCache{
		name:   name,
		client: client,
		config: config,
	}
}

// Name returns the plugin name
func (g *GormCache) Name() string {
	return g.name
}

// Initialize initializes the plugin
func (g *GormCache) Initialize(db *gorm.DB) error {
	return db.Callback().Query().Replace("gorm:query", g.queryCallback)
}

// queryCallback is a callback function for query operations
func (g *GormCache) queryCallback(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	enableCache := g.enableCache(db)

	// build query sql
	callbacks.BuildQuerySQL(db)
	if db.DryRun || db.Error != nil {
		return
	}

	var (
		key string
		err error
		hit bool
	)
	if enableCache {
		key = g.cacheKey(db)

		// get value from cache
		hit, err = g.loadCache(db, key)
		if err != nil {
			log.Printf("*** load cache failed, err: '%v', hit value: %v", err, hit)
			return
		}

		// hit cache
		if hit {
			return
		}

		// cache miss, continue database operation
		//log.Printf("------------------------- miss cache, key: %v", key)
	}

	if !hit {
		g.queryDB(db)

		if enableCache {
			if err = g.setCache(db, key); err != nil {
				log.Printf("*** set cache failed: %v", err)
			}
		}
	}
}

func (g *GormCache) enableCache(db *gorm.DB) bool {
	ctx := db.Statement.Context

	// check if use cache
	useCache, ok := ctx.Value(UseCacheKey).(bool)
	if !ok || !useCache {
		return false // do not use cache, skip this callback
	}
	return true
}

func (g *GormCache) loadCache(db *gorm.DB, key string) (bool, error) {
	value, err := g.client.Get(db.Statement.Context, key)
	if err != nil {
		return false, err
	}

	if value == nil {
		return false, nil
	}

	// cache hit, scan value to destination
	if err = json.Unmarshal(value.([]byte), &db.Statement.Dest); err != nil {
		return false, err
	}
	db.RowsAffected = int64(db.Statement.ReflectValue.Len())
	return true, nil
}

func (g *GormCache) setCache(db *gorm.DB, key string) error {
	ctx := db.Statement.Context

	// get cache ttl from context or config
	ttl, ok := ctx.Value(CacheTTLKey).(time.Duration)
	if !ok {
		ttl = g.config.TTL // use default ttl
	}
	//log.Printf("ttl: %v", ttl)

	// set value to cache with ttl
	return g.client.Set(ctx, key, db.Statement.Dest, ttl)
}

func (g *GormCache) queryDB(db *gorm.DB) {
	rows, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
	if err != nil {
		db.AddError(err)
		return
	}
	defer func() {
		db.AddError(rows.Close())
	}()
	gorm.Scan(rows, db, 0)
}

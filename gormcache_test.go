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

package gormcache

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	memcache "github.com/bradfitz/gomemcache/memcache"
	redis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TestUserRedis struct {
	ID   int
	Name string
}

type TestUserMC struct {
	ID   int
	Name string
}

type TestUserBoltDB struct {
	ID   int
	Name string
}

var (
	dbRedis   *gorm.DB
	dbMC      *gorm.DB
	dbBoltDB  *gorm.DB
	rdb       *redis.Client
	mdb       *memcache.Client
	bdb       *bolt.DB
	userCount = 100
)

func init() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")
	//dsn := fmt.Sprintf("host='%v' port='%v' user='%v'  password='%v' dbname='%v' sslmode=disable", dbHost, dbPort, dbUser, dbPwd, dbName)
	dsn := fmt.Sprintf("%v:%v@(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPwd, dbHost, dbPort, dbName)

	dbRedis, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	dbRedis.Migrator().DropTable(TestUserRedis{})

	dbRedis.AutoMigrate(TestUserRedis{})

	for i := 0; i < userCount; i++ {
		if err = dbRedis.Save(&TestUserRedis{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Fatalln(err)
		}
	}

	dbMC, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	dbMC.Migrator().DropTable(TestUserMC{})

	dbMC.AutoMigrate(TestUserMC{})

	for i := 0; i < userCount; i++ {
		if err = dbMC.Save(&TestUserMC{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Fatalln(err)
		}
	}

	dbBoltDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	dbBoltDB.Migrator().DropTable(TestUserBoltDB{})

	dbBoltDB.AutoMigrate(TestUserBoltDB{})

	for i := 0; i < userCount; i++ {
		if err = dbMC.Save(&TestUserBoltDB{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Fatalln(err)
		}
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
	})

	mdb = memcache.New("127.0.0.1:11211")
}

// TestRedisCache tests the cache plugin functionality
func TestRedisCache(t *testing.T) {
	var err error

	args := []struct {
		UseCache bool
		TTL      time.Duration
		ID       int
	}{
		{
			UseCache: false,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       10,
		},
		{
			UseCache: true,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       5,
		},
		{
			UseCache: true,
			ID:       15,
		},
		{
			UseCache: true,
			TTL:      10 * time.Second,
			ID:       10,
		},
	}

	cache := NewGormCache("my_cache", NewRedisClient(rdb), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = dbRedis.Use(cache)
	assert.NoError(t, err)

	for _, arg := range args {
		var users []TestUserRedis
		ctx := context.WithValue(context.Background(), UseCacheKey, true)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, CacheTTLKey, arg.TTL)
		}

		// query with cache and custom ttl
		err = dbRedis.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

// TestMemcacheCache tests the cache plugin functionality
func TestMemcacheCache(t *testing.T) {
	var err error

	args := []struct {
		UseCache bool
		TTL      time.Duration
		ID       int
	}{
		{
			UseCache: false,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       10,
		},
		{
			UseCache: true,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       5,
		},
		{
			UseCache: true,
			ID:       15,
		},
		{
			UseCache: true,
			TTL:      10 * time.Second,
			ID:       10,
		},
	}

	cache := NewGormCache("my_cache", NewMemcacheClient(mdb), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = dbMC.Use(cache)
	assert.NoError(t, err)

	for _, arg := range args {
		var users []TestUserMC
		ctx := context.WithValue(context.Background(), UseCacheKey, true)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, CacheTTLKey, arg.TTL)
		}

		// query with cache and custom ttl
		err = dbMC.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

// TestBoltDBCache tests the cache plugin functionality
func TestBoltDBCache(t *testing.T) {
	var err error

	args := []struct {
		UseCache bool
		TTL      time.Duration
		ID       int
	}{
		{
			UseCache: false,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       10,
		},
		{
			UseCache: true,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       5,
		},
		{
			UseCache: true,
			ID:       15,
		},
		{
			UseCache: true,
			TTL:      10 * time.Second,
			ID:       10,
		},
	}

	bdb, err := bolt.Open("/tmp/cache.db", 0600, nil)
	if err != nil {
		log.Fatalf("could not open db, %v", err)
	}
	defer bdb.Close()
	bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			log.Fatalf("could not create root bucket: %v", err)
		}
		return nil
	})

	cache := NewGormCache("my_cache", NewBboltClient(bdb), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = dbBoltDB.Use(cache)
	assert.NoError(t, err)

	for _, arg := range args {
		var users []TestUserBoltDB
		ctx := context.WithValue(context.Background(), UseCacheKey, true)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, CacheTTLKey, arg.TTL)
		}

		// query with cache and custom ttl
		err = dbBoltDB.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

// BenchmarkCache benchmarks the cache plugin performance
func BenchmarkRedisCache(b *testing.B) {
	cache := NewGormCache("my_cache", NewRedisClient(rdb), CacheConfig{
		TTL:    10 * time.Second,
		Prefix: "cache:",
	})
	dbRedis.Use(cache)

	var users []TestUserRedis

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dbRedis.Session(&gorm.Session{Context: context.WithValue(context.Background(), UseCacheKey, true)}).Where("id > ?", 10).Find(&users)
	}
}

// BenchmarkMemcachedCache benchmarks the cache plugin performance
func BenchmarkMemcachedCache(b *testing.B) {
	cache := NewGormCache("my_cache", NewMemcacheClient(mdb), CacheConfig{
		TTL:    10 * time.Second,
		Prefix: "cache:",
	})
	dbMC.Use(cache)

	var users []TestUserMC

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dbMC.Session(&gorm.Session{Context: context.WithValue(context.Background(), UseCacheKey, true)}).Where("id > ?", 10).Find(&users)
	}
}

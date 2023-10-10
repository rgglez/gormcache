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

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestUser struct {
	ID   int
	Name string
}

var (
	db        *gorm.DB
	rdb       *redis.Client
	userCount = 100
)

func init() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("host='%v' port='%v' user='%v'  password='%v' dbname='%v' sslmode=disable", dbHost, dbPort, dbUser, dbPwd, dbName)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	db.Migrator().DropTable(TestUser{})

	db.AutoMigrate(TestUser{})

	for i := 0; i < userCount; i++ {
		if err = db.Save(&TestUser{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Fatalln(err)
		}
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
	})
}

// TestCache tests the cache plugin functionality
func TestCache(t *testing.T) {
	var err error

	cache := NewGormCache("my_cache", NewRedisClient(rdb), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = db.Use(cache)
	assert.NoError(t, err)

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

	for _, arg := range args {
		var users []TestUser
		ctx := context.WithValue(context.Background(), UseCacheKey, arg.UseCache)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, CacheTTLKey, arg.TTL)
		}

		// query with cache and custom ttl
		err = db.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

// BenchmarkCache benchmarks the cache plugin performance
func BenchmarkCache(b *testing.B) {
	cache := NewGormCache("my_cache", NewRedisClient(rdb), CacheConfig{
		TTL:    10 * time.Second,
		Prefix: "cache:",
	})
	db.Use(cache)

	var users []TestUser

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Session(&gorm.Session{Context: context.WithValue(context.Background(), UseCacheKey, true)}).Where("id > ?", 10).Find(&users)
	}
}

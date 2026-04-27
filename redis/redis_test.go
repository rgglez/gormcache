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

package gormcacheredis_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	gormcache "github.com/rgglez/gormcache"
	gormcacheredis "github.com/rgglez/gormcache/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TestUserRedis struct {
	ID   int
	Name string
}

var (
	dbRedis   *gorm.DB
	rdb       *redis.Client
	userCount = 100
)

func init() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return
	}
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("%v:%v@(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPwd, dbHost, dbPort, dbName)

	var err error
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

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
	})
}

func TestRedisCache(t *testing.T) {
	if dbRedis == nil {
		t.Skip("DB_HOST not set, skipping integration test")
	}

	args := []struct {
		UseCache bool
		TTL      time.Duration
		ID       int
	}{
		{UseCache: false, ID: 10},
		{UseCache: true, TTL: 5 * time.Second, ID: 10},
		{UseCache: true, ID: 10},
		{UseCache: true, TTL: 5 * time.Second, ID: 5},
		{UseCache: true, ID: 15},
		{UseCache: true, TTL: 10 * time.Second, ID: 10},
	}

	cache := gormcache.NewGormCache("my_cache", gormcacheredis.NewRedisClient(rdb), gormcache.CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err := dbRedis.Use(cache)
	assert.NoError(t, err)

	for _, arg := range args {
		var users []TestUserRedis
		ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, gormcache.CacheTTLKey, arg.TTL)
		}
		err = dbRedis.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

func BenchmarkRedisCache(b *testing.B) {
	if dbRedis == nil {
		b.Skip("DB_HOST not set, skipping integration benchmark")
	}

	cache := gormcache.NewGormCache("my_cache", gormcacheredis.NewRedisClient(rdb), gormcache.CacheConfig{
		TTL:    10 * time.Second,
		Prefix: "cache:",
	})
	dbRedis.Use(cache)

	var users []TestUserRedis
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbRedis.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, true)}).Where("id > ?", 10).Find(&users)
	}
}

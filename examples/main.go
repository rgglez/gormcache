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

package main

import (
	"context"
	"log"
	"time"

	redis "github.com/go-redis/redis/v8"
	gormcache "github.com/rgglez/gormcache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host='0.0.0.0' port='5432' user='evan' dbname='cache_test' password='' sslmode=disable TimeZone=Asia/Shanghai"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.Default(),
			logger.Config{
				LogLevel: logger.Info,
				Colorful: true,
			},
		),
	})

	// User is a sample model
	type User struct {
		ID   int
		Name string
	}

	db.AutoMigrate(User{})

	rdb := redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "123456",
	})

	cache := gormcache.NewGormCache("my_cache", gormcache.NewRedisClient(rdb), gormcache.CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})

	if err := db.Use(cache); err != nil {
		log.Fatal(err)
	}

	var users []User
	/*
		// mock data
		for i := 0; i < 100; i++ {
			db.Save(&User{Name: fmt.Sprintf("%X", byte('A'+i))})
		}
	*/

	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, true)}).
		Where("id > ?", 10).Find(&users) // use cache with default ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), gormcache.UseCacheKey, true), gormcache.CacheTTLKey, 10*time.Second)}).
		Where("id > ?", 5).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), gormcache.UseCacheKey, true), gormcache.CacheTTLKey, 20*time.Second)}).
		Where("id > ?", 5).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, false)}).
		Where("id > ?", 10).Find(&users) // do not use cache
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), gormcache.UseCacheKey, true), gormcache.CacheTTLKey, 10*time.Second)}).
		Where("id > ?", 10).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)
}

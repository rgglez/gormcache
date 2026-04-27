/*
   Copyright 2023 Rodolfo González González

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

package gormcachebbolt_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormcache "github.com/rgglez/gormcache"
	gormcachebbolt "github.com/rgglez/gormcache/bbolt"
	bolt "go.etcd.io/bbolt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TestUserBoltDB struct {
	ID   int
	Name string
}

var (
	dbBoltDB  *gorm.DB
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
	dbBoltDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	dbBoltDB.Migrator().DropTable(TestUserBoltDB{})
	dbBoltDB.AutoMigrate(TestUserBoltDB{})

	for i := 0; i < userCount; i++ {
		if err = dbBoltDB.Save(&TestUserBoltDB{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Fatalln(err)
		}
	}
}

func TestBoltDBCache(t *testing.T) {
	if dbBoltDB == nil {
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

	bdb, err := bolt.Open("/tmp/cache_bbolt_test.db", 0600, nil)
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

	cache := gormcache.NewGormCache("my_cache", gormcachebbolt.NewBboltClient(bdb), gormcache.CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = dbBoltDB.Use(cache)
	assert.NoError(t, err)

	for _, arg := range args {
		var users []TestUserBoltDB
		ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, gormcache.CacheTTLKey, arg.TTL)
		}
		err = dbBoltDB.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

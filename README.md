# gorm-cache

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub all releases](https://img.shields.io/github/downloads/rgglez/gorm-cache/total)
![GitHub issues](https://img.shields.io/github/issues/rgglez/gorm-cache)
![GitHub commit activity](https://img.shields.io/github/commit-activity/y/rgglez/gorm-cache)

[![Go Report Card](https://goreportcard.com/badge/github.com/rgglez/gorm-cache)](https://goreportcard.com/report/github.com/rgglez/gorm-cache)
[![GitHub release](https://img.shields.io/github/release/rgglez/gorm-cache.svg)](https://github.com/rgglez/gorm-cache/releases/)


gorm-cache is a fork of the [evangwt/grc](https://github.com/evangwt/grc) [GORM](https://gorm.io/index.html) plugin that provides a way to cache data using redis and memcached at the moment.

This fork separates the cache backend specifics to their own files in the same gormcache package, and adds the memcached backend.

## Features

- Easy to use: just add gorm-cache as a GORM plugin and use GORM session options to control the cache behavior.
- Flexible to customize: you can configure the cache prefix, ttl, and redis or memcached client according to your needs.

## Installation

### Dependencies

* [GORM](https://gorm.io/index.html)

```bash
go get -u gorm.io/gorm
```

* [redis library v8](https://github.com/redis/go-redis)

```bash
go get -u github.com/go-redis/redis/v8
```

, or

* [memcached library](https://github.com/bradfitz/gomemcache)

```bash
go get github.com/bradfitz/gomemcache/memcache
```

Then you can install gorm-cache using go get:

```bash
go get -u github.com/rgglez/gorm-cache
```

## Usage

To use gorm-cache, you need to create a gorm-cache instance with a redis or memcached client and a cache config, and then add it as a GORM plugin. For example:

```go
package main

import (
        "github.com/rgglez/gorm-cache"
        "github.com/go-redis/redis/v8"
        "gorm.io/driver/postgres"
        "gorm.io/gorm"
)

func main() {
        // connect to postgres database
        dsn := "host='0.0.0.0' port='5432' user='evan' dbname='cache_test' password='' sslmode=disable TimeZone=Asia/Shanghai"
        db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

        // connect to redis database
        rdb := redis.NewClient(&redis.Options{
                Addr:     "0.0.0.0:6379",
                Password: "123456",
        })

        // create a gorm cache instance with a redis client and a cache config
        cache := gormcache.NewGormCache("my_cache", gormcache.NewRedisClient(rdb), gormcache.CacheConfig{
                TTL:    60 * time.Second,
                Prefix: "cache:",
        })

        // add the gorm cache instance as a gorm plugin
        if err := db.Use(cache); err != nil {
                log.Fatal(err)
        }

        // now you can use gorm session options to control the cache behavior
}
```

To enable or disable the cache for a query, you can use the `gormcache.UseCacheKey` context value with a boolean value. For example:

```go
// use cache with default ttl
db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, true)}).
                Where("id > ?", 10).Find(&users)

// do not use cache
db.Session(&gorm.Session{Context: context.WithValue(context.Background(), gormcache.UseCacheKey, false)}).
                Where("id > ?", 10).Find(&users)
```

To set a custom ttl for a query, you can use the `gormcache.CacheTTLKey` context value with a time.Duration value. For example:

```go
// use cache with custom ttl
db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), gormcache.UseCacheKey, true), gormcache.CacheTTLKey, 10*time.Second)}).
                Where("id > ?", 5).Find(&users)
```

For more examples and details, please refer to the [example code](https://github.com/rgglez/gorm-cache/tree/main/example).

## License

Read the [LICENSE](https://github.com/rgglez/gorm-cache/blob/main/LICENSE) file for more information.

This module is based on [evangwt/grc](https://github.com/evangwt/grc).

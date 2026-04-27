# gormcache

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub all releases](https://img.shields.io/github/downloads/rgglez/gormcache/total)
![GitHub issues](https://img.shields.io/github/issues/rgglez/gormcache)
![GitHub commit activity](https://img.shields.io/github/commit-activity/y/rgglez/gormcache)
[![Go Report Card](https://goreportcard.com/badge/github.com/rgglez/gormcache)](https://goreportcard.com/report/github.com/rgglez/gormcache)
[![GitHub release](https://img.shields.io/github/release/rgglez/gormcache.svg)](https://github.com/rgglez/gormcache/releases/)
![GitHub stars](https://img.shields.io/github/stars/rgglez/gormcache?style=social)
![GitHub forks](https://img.shields.io/github/forks/rgglez/gormcache?style=social)

gormcache is a [GORM](https://gorm.io/index.html) plugin that caches query results using Redis, BoltDB, or Memcached as backend. It is a fork of [evangwt/grc](https://github.com/evangwt/grc) with added backends and a monorepo structure for independent versioning.

## Repository structure

This is a Go monorepo. Each backend is a separate module with its own versioning:

| Module | Import path | Latest |
|--------|-------------|--------|
| Core plugin | `github.com/rgglez/gormcache` | `v0.0.16` |
| Redis backend | `github.com/rgglez/gormcache/redis` | `v0.1.0` |
| BoltDB backend | `github.com/rgglez/gormcache/bbolt` | `v0.1.0` |
| Memcached backend | `github.com/rgglez/gormcache/memcached` | `v0.1.0` |

The core module defines the `CacheClient` interface and the `GormCache` plugin. Backend modules are optional — only install the one you need.

## Features

- Easy to use: add gormcache as a GORM plugin, control caching per query via context values.
- Flexible: configure prefix, TTL, and backend independently.
- Modular: only import the backend you need — no unused dependencies.

## Installation

Always install the core plus one backend:

### Redis

```bash
go get github.com/rgglez/gormcache@latest
go get github.com/rgglez/gormcache/redis@latest
```

### BoltDB

```bash
go get github.com/rgglez/gormcache@latest
go get github.com/rgglez/gormcache/bbolt@latest
```

### Memcached

```bash
go get github.com/rgglez/gormcache@latest
go get github.com/rgglez/gormcache/memcached@latest
```

## Usage

### Redis

```go
package main

import (
    "context"
    "log"
    "time"

    redis "github.com/redis/go-redis/v9"
    gormcache "github.com/rgglez/gormcache"
    gormcacheredis "github.com/rgglez/gormcache/redis"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    db, _ := gorm.Open(postgres.Open("...dsn..."), &gorm.Config{})

    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
    })

    cache := gormcache.NewGormCache("my_cache", gormcacheredis.NewRedisClient(rdb), gormcache.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })
    if err := db.Use(cache); err != nil {
        log.Fatal(err)
    }

    var users []User
    ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
}
```

### BoltDB

```go
package main

import (
    "context"
    "log"
    "time"

    bolt "go.etcd.io/bbolt"
    gormcache "github.com/rgglez/gormcache"
    gormcachebbolt "github.com/rgglez/gormcache/bbolt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    db, _ := gorm.Open(mysql.Open("...dsn..."), &gorm.Config{})

    bdb, err := bolt.Open("/tmp/cache.db", 0600, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer bdb.Close()
    bdb.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("DB"))
        return err
    })

    cache := gormcache.NewGormCache("my_cache", gormcachebbolt.NewBboltClient(bdb), gormcache.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })
    if err := db.Use(cache); err != nil {
        log.Fatal(err)
    }

    var users []User
    ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
}
```

> **Note:** BoltDB has no native TTL support. Entries persist until the database file is deleted or you implement manual eviction.

### Memcached

```go
package main

import (
    "context"
    "log"
    "time"

    memcache "github.com/bradfitz/gomemcache/memcache"
    gormcache "github.com/rgglez/gormcache"
    gormcachememcached "github.com/rgglez/gormcache/memcached"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    db, _ := gorm.Open(mysql.Open("...dsn..."), &gorm.Config{})

    mdb := memcache.New("127.0.0.1:11211")

    cache := gormcache.NewGormCache("my_cache", gormcachememcached.NewMemcacheClient(mdb), gormcache.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })
    if err := db.Use(cache); err != nil {
        log.Fatal(err)
    }

    var users []User
    ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
}
```

## Controlling cache per query

Enable or disable caching and set a custom TTL via context values:

```go
// Use cache with default TTL (from CacheConfig)
ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, true)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)

// Use cache with custom TTL
ctx := context.WithValue(
    context.WithValue(context.Background(), gormcache.UseCacheKey, true),
    gormcache.CacheTTLKey, 10*time.Second,
)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 5).Find(&users)

// Bypass cache
ctx := context.WithValue(context.Background(), gormcache.UseCacheKey, false)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
```

## Migration guide from v0.0.15

Starting with `v0.0.16`, backend clients are in separate modules. The core API is unchanged.

### 1. Install the new backend module

```bash
# Choose the backend you were using
go get github.com/rgglez/gormcache/redis@latest
# or: go get github.com/rgglez/gormcache/bbolt@latest
# or: go get github.com/rgglez/gormcache/memcached@latest
```

### 2. Add the backend import

**Before:**
```go
import gormcache "github.com/rgglez/gormcache"
```

**After (Redis example):**
```go
import (
    gormcache      "github.com/rgglez/gormcache"
    gormcacheredis "github.com/rgglez/gormcache/redis"
)
```

### 3. Update constructor calls

The constructors moved to the backend packages. The function signatures are identical.

| Before | After |
|--------|-------|
| `gormcache.NewRedisClient(rdb)` | `gormcacheredis.NewRedisClient(rdb)` |
| `gormcache.NewBboltClient(bdb)` | `gormcachebbolt.NewBboltClient(bdb)` |
| `gormcache.NewMemcacheClient(mdb)` | `gormcachememcached.NewMemcacheClient(mdb)` |

`NewGormCache`, `CacheConfig`, `UseCacheKey`, `CacheTTLKey` — all remain in `github.com/rgglez/gormcache` and are **unchanged**.

### 4. Update the Redis client library (if using Redis)

The old example used `github.com/go-redis/redis/v8`. The Redis backend now uses `github.com/redis/go-redis/v9`. The API is nearly identical; the main difference is the import path:

```bash
go get github.com/redis/go-redis/v9
```

```go
// Before
import redis "github.com/go-redis/redis/v8"

// After
import redis "github.com/redis/go-redis/v9"
```

### Complete before/after example

**Before (v0.0.15):**
```go
import (
    gormcache "github.com/rgglez/gormcache"
    redis "github.com/go-redis/redis/v8"
)

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
cache := gormcache.NewGormCache("my_cache", gormcache.NewRedisClient(rdb), gormcache.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

**After (v0.0.16+):**
```go
import (
    gormcache      "github.com/rgglez/gormcache"
    gormcacheredis "github.com/rgglez/gormcache/redis"
    redis          "github.com/redis/go-redis/v9"
)

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
cache := gormcache.NewGormCache("my_cache", gormcacheredis.NewRedisClient(rdb), gormcache.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

## Makefile

A `Makefile` is included for common monorepo tasks. Run `make help` to see all targets.

### Tagging

`VERSION` is required for all tag targets.

| Rule | Example | Description |
|------|---------|-------------|
| `tag-core` | `make tag-core VERSION=v0.0.17` | Tag the core module |
| `tag-redis` | `make tag-redis VERSION=v0.1.1` | Tag the Redis plugin (creates `redis/v0.1.1`) |
| `tag-bbolt` | `make tag-bbolt VERSION=v0.1.1` | Tag the BoltDB plugin |
| `tag-memcached` | `make tag-memcached VERSION=v0.1.1` | Tag the Memcached plugin |
| `tag-plugins` | `make tag-plugins VERSION=v0.1.1` | Tag all three plugins with the same version |
| `push-tags` | `make push-tags` | Push all local tags to origin |

### Dependency management

| Rule | Example | Description |
|------|---------|-------------|
| `update-core` | `make update-core VERSION=v0.0.17` | Run `go get core@VERSION` in each plugin module |
| `tidy` | `make tidy` | Run `go mod tidy` in all modules |

### Verification

| Rule | Description |
|------|-------------|
| `build` | `go build ./...` in all modules |
| `vet` | `go vet ./...` in all modules |
| `test` | `go test ./...` in all modules |
| `versions` | Show the latest tag for each module |

## License

Apache-2.0 license. See [LICENSE](https://github.com/rgglez/gormcache/blob/main/LICENSE).

This module is based on [evangwt/grc](https://github.com/evangwt/grc).

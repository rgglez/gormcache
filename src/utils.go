package gormcache

import (
	"crypto/sha256"
	"encoding/hex"

	"gorm.io/gorm"
)

func (g *GormCache) cacheKey(db *gorm.DB) string {
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	hash := sha256.Sum256([]byte(sql))
	key := g.config.Prefix + hex.EncodeToString(hash[:])
	//log.Printf("key: %v, sql: %v", key, sql)
	return key
}

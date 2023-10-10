module gormcache-test

go 1.21.2

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/stretchr/testify v1.8.4
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5

)

require github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rgglez/gorm-cache v0.0.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rgglez/gorm-cache => ../src

module tests

go 1.21

//replace github.com/jefferyjob/go-redislock => ..

require (
	github.com/jefferyjob/go-redislock v1.7.0-beta.2
	github.com/jefferyjob/go-redislock/adapter/go-redis/V9 v0.0.0-20251210060753-5b2e8d62842e
	github.com/redis/go-redis/v9 v9.17.2
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

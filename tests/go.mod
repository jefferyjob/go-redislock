module github.com/jefferyjob/go-redislock/tests

go 1.24.5

replace github.com/jefferyjob/go-redislock@latest => ..

require (
	github.com/jefferyjob/go-redislock v1.5.0
	github.com/jefferyjob/go-redislock/adapter/go-redis/v9 v9.0.0-20251023074556-1d78f01c9a96
	github.com/redis/go-redis/v9 v9.14.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

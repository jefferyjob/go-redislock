module github.com/jefferyjob/go-redislock/tests

go 1.24.5

toolchain go1.24.10

//replace github.com/jefferyjob/go-redislock@latest => ..

require (
	github.com/jefferyjob/go-redislock v1.6.1-0.20251122070819-0e374055d35c
	github.com/jefferyjob/go-redislock/adapter/go-redis/v9 v9.0.0-20251023074556-1d78f01c9a96
	github.com/redis/go-redis/v9 v9.17.0
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

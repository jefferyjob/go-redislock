module github.com/jefferyjob/go-redislock/examples/adapter/go-redis/v9

go 1.24.5

// For local testing, please ignore this configuration
replace github.com/jefferyjob/go-redislock@latest => ../../..

require (
	github.com/jefferyjob/go-redislock v1.4.0
	github.com/jefferyjob/go-redislock/adapter/go-redis/v9 v9.0.0-20251022110012-7fa8953ae22a
	github.com/redis/go-redis/v9 v9.14.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
)

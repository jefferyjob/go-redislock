module github.com/jefferyjob/go-redislock/adapter/go-redis/V8

go 1.21

replace github.com/jefferyjob/go-redislock => ../../..

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/jefferyjob/go-redislock v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
)

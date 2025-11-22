module github.com/jefferyjob/go-redislock/adapter/go-redis/V9

go 1.21

replace github.com/jefferyjob/go-redislock => ../../..

require (
	github.com/jefferyjob/go-redislock v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.17.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
)

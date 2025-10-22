module demo

go 1.24.5

require (
	github.com/jefferyjob/go-redislock v1.4.1-0.20251022084857-ae8992cf65aa
	github.com/jefferyjob/go-redislock/adapter/go-redis/v9 v9.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.14.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/jefferyjob/go-redislock => ../..

replace github.com/jefferyjob/go-redislock/adapter/go-redis/v9 => ../../adapter/go-redis/v9

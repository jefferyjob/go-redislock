module github.com/jefferyjob/go-redislock/adapter/go-redis/V7

go 1.21

replace github.com/jefferyjob/go-redislock => ../../..

require (
	github.com/go-redis/redis/v7 v7.4.1
	github.com/jefferyjob/go-redislock v0.0.0-00010101000000-000000000000
)

require github.com/google/uuid v1.6.0 // indirect

module github.com/jefferyjob/go-redislock/adapter/go-redis/V7

go 1.24.5

replace github.com/jefferyjob/go-redislock => ../../..

require (
	github.com/go-redis/redis/v7 v7.4.1
	github.com/jefferyjob/go-redislock v1.3.0
)

require github.com/google/uuid v1.6.0 // indirect

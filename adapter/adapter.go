package adapter

import (
	"fmt"
	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	gfRdbV1 "github.com/gogf/gf/database/gredis"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
	adapterGfV1 "github.com/jefferyjob/go-redislock/adapter/gf/v1"
	adapterGfV2 "github.com/jefferyjob/go-redislock/adapter/gf/v2"
	adapterGz "github.com/jefferyjob/go-redislock/adapter/gozero"
	adapterV7 "github.com/jefferyjob/go-redislock/adapter/v7"
	adapterV8 "github.com/jefferyjob/go-redislock/adapter/v8"
	adapterV9 "github.com/jefferyjob/go-redislock/adapter/v9"
	v9 "github.com/redis/go-redis/v9"
	gzRdb "github.com/zeromicro/go-zero/core/stores/redis"
	"log"
	"runtime/debug"
)

// New Create a new Redis adapter based on the original Redis client
func New(rawClient interface{}) (redislock.RedisInter, error) {
	switch client := rawClient.(type) {
	case *v7.Client:
		return adapterV7.New(client), nil
	case *v8.Client:
		return adapterV8.New(client), nil
	case *v9.Client:
		return adapterV9.New(client), nil
	case *gzRdb.Redis:
		return adapterGz.New(client), nil
	case *gfRdbV1.Redis:
		return adapterGfV1.New(client), nil
	case *gfRdbV2.Redis:
		return adapterGfV2.New(client), nil
	default:
		return nil, fmt.Errorf("unsupported redis client type: %T", rawClient)
	}
}

// MustNew Create a new Redis adapter based on the original Redis client
func MustNew(rawClient interface{}) redislock.RedisInter {
	adapter, err := New(rawClient)
	if err != nil {
		msg := fmt.Sprintf("%+v\n\n%s", err.Error(), debug.Stack())
		log.Println(msg)
		panic(msg)
	}
	return adapter
}

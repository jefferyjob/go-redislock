package tests

import (
	"fmt"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v9"
	"github.com/redis/go-redis/v9"
)

const (
	// 默认锁超时时间
	lockTime = 5 * time.Second
	// 默认请求超时时间
	requestTimeout = lockTime
)

var (
	addr = "127.0.0.1"
	port = "63790"
	// luaSetScript = `return redis.call("SET", KEYS[1], ARGV[1])`
	// luaGetScript = `return redis.call("GET", KEYS[1])`
	// luaDelScript = `return redis.call("DEL", KEYS[1])`
)

// Redis服务器测试
// 下面的代码将借助 redis 服务器进行测试，可以更加方便的测试服务中的问题
// 你可以实用下面的命令启动一个redis容器进行测试
// docker run -d -p 63790:6379 --name go_redis_lock redis
func getRedisClient() redislock.RedisInter {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	})
	return adapter.New(rdb)
}

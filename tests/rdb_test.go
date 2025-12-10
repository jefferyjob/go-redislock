package tests

import (
	"fmt"
	"sync"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/V9"
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

	once       sync.Once
	redisInter redislock.RedisInter
)

// Redis 服务器集成测试
//
// 本测试依赖实际的 Redis 服务，用于验证分布式锁在真实环境下的行为。
// 你可以通过以下命令快速启动一个本地 Redis 容器：
//
//	docker run -d -p 63790:6379 --name go_redis_lock redis
//
// 运行后，测试代码将自动连接到该容器的 Redis 实例，
// 以便更方便地调试和定位服务中的潜在问题。
func getRedisClient() redislock.RedisInter {
	once.Do(func() {
		rdb := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", addr, port),
		})
		redisInter = adapter.New(rdb)
	})
	return redisInter
}

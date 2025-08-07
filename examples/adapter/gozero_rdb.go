package adapter

import (
	"context"
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/jefferyjob/go-redislock/adapter/gozero"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
)

// GoZeroRedisLock 演示如何在 gozero 框架的 redis 客户端上使用 redislock 库
func GoZeroRedisLock() {
	// Initialize redis adapter (only once)
	rdbAdapter := gozero.New(zeroRdb.MustNewRedis(zeroRdb.RedisConf{
		Host: "localhost:6379",
		Type: "node",
	}))

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(rdbAdapter, "test_key")

	// acquire lock
	err := lock.Lock(ctx)
	if err != nil {
		fmt.Println("lock acquisition failed：", err)
		return
	}
	defer lock.UnLock(ctx) // unlock

	// Perform tasks during lockdown
	// ...

	fmt.Println("task execution completed")
}

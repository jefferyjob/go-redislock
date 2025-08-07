package adapter

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2" // 注册 Redis 适配器（必须）
	"github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/jefferyjob/go-redislock/adapter/gf/v2"
)

// 演示如何在 gf 框架的 gredis v2 客户端上使用 redislock 库
func demoGfV2() {
	// Initialize redis (only once)
	rdb, err := gredis.New(&gredis.Config{
		Address: "localhost:6379",
	})
	if err != nil {
		fmt.Println("failed to create redis client:", err)
		return
	}

	// Create a Redis client using GfV2
	rdbAdapter := v2.New(rdb)

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(rdbAdapter, "test_key")

	// acquire lock
	err = lock.Lock(ctx)
	if err != nil {
		fmt.Println("lock acquisition failed：", err)
		return
	}
	defer lock.UnLock(ctx) // unlock

	// Perform tasks during lockdown
	// ...

	fmt.Println("task execution completed")
}

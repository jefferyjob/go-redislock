package adapter

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2" // 注册 Redis 适配器（必须）
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
)

// GfV2RdbLock 演示如何在 gf 框架的 gredis v2 客户端上使用 redislock 库
func GfV2RdbLock() {
	// Initialize redis (only once)
	rdb, err := gfRdbV2.New(&gfRdbV2.Config{
		Address: "localhost:6379",
	})
	if err != nil {
		fmt.Println("failed to create redis client:", err)
		return
	}

	// Create a Redis client using GfV2
	redisClient := redislock.NewGfRedisV2Adapter(rdb)

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(redisClient, "test_key")

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

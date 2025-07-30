package adapter

import (
	"context"
	"fmt"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	gV2 "github.com/gogf/gf/v2/frame/g"
	redislock "github.com/jefferyjob/go-redislock"
)

func initRedis() error {
	_, err := gfRdbV2.New(&gfRdbV2.Config{
		Address: "localhost:6379",
	})
	return err
}

// GfV2RdbLock 演示如何在 gf 框架的 gredis v2 客户端上使用 redislock 库
func GfV2RdbLock() {
	// Initialize Redis (only once)
	if err := initRedis(); err != nil {
		fmt.Println("failed to init redis client:", err)
		return
	}

	// Create a Redis client using GfV2
	redisClient := redislock.NewGfRedisV2Adapter(gV2.Redis())

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(redisClient, "test_key")

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

package redis_client_adapter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	redislock "github.com/jefferyjob/go-redislock"
)

// RedisV8Lock 演示如何在官方 go-redis v8 客户端上使用 redislock 库
func RedisV8Lock() {
	redisClient := redislock.NewRedisV8Adapter(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}))

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(ctx, redisClient, "test_key")

	// acquire lock
	err := lock.Lock()
	if err != nil {
		fmt.Println("lock acquisition failed：", err)
		return
	}
	defer lock.UnLock() // unlock

	// Perform tasks during lockdown
	// ...

	fmt.Println("task execution completed")
}

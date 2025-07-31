package adapter

import (
	"context"
	"fmt"
	v7 "github.com/go-redis/redis/v7"
	redislock "github.com/jefferyjob/go-redislock"
)

// RedisV7Lock 演示如何在官方 go-redis v7 客户端上使用 redislock 库
func RedisV7Lock() {
	// Initialize redis adapter (only once)
	redisClient := redislock.NewRedisV7Adapter(v7.NewClient(&v7.Options{
		Addr: "localhost:6379",
	}))

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

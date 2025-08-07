package adapter

import (
	"context"
	"fmt"
	v8 "github.com/go-redis/redis/v8"
	redislock "github.com/jefferyjob/go-redislock"
)

// RedisV8Lock 演示如何在官方 go-redis v8 客户端上使用 redislock 库
func RedisV8Lock() {
	// Initialize redis adapter (only once)
	redisClient := redislock.NewRedisV8Adapter(v8.NewClient(&v8.Options{
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

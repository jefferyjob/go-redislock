package adapter

import (
	"context"
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	v9 "github.com/redis/go-redis/v9"
)

// RedisV9Lock 演示如何在官方 go-redis v9 客户端上使用 redislock 库
func RedisV9Lock() {
	redisClient := redislock.NewRedisV9Adapter(v9.NewClient(&v9.Options{
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

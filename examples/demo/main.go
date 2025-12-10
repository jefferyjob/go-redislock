package main

import (
	"context"
	"fmt"

	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/V9"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Create a Redis client adapter
	rdbAdapter := adapter.New(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
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

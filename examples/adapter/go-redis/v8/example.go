package v8

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v8"
)

func main() {
	// Initialize redis adapter (only once)
	redisClient := adapter.New(redis.NewClient(&redis.Options{
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

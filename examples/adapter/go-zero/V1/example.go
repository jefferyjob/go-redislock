package V1

import (
	"context"
	"fmt"

	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-zero/V1"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func main() {
	// Initialize redis adapter (only once)
	rdbAdapter := adapter.New(redis.MustNewRedis(redis.RedisConf{
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

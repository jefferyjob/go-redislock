package adapter

// GoZeroRedisLock 演示如何在 gozero 框架的 redis 客户端上使用 redislock 库
// func GoZeroRedisLock() {
// 	redisClient := redislock.NewGoZeroRdbAdapter(redis.MustNewRedis(redis.RedisConf{
// 		Host: "localhost:6379",
// 	}))
//
// 	// Create a context for canceling lock operations
// 	ctx := context.Background()
//
// 	// Create a RedisLock object
// 	lock := redislock.New(redisClient, "test_key")
//
// 	// acquire lock
// 	err := lock.Lock(ctx)
// 	if err != nil {
// 		fmt.Println("lock acquisition failed：", err)
// 		return
// 	}
// 	defer lock.UnLock(ctx) // unlock
//
// 	// Perform tasks during lockdown
// 	// ...
//
// 	fmt.Println("task execution completed")
// }

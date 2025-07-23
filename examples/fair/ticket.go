package fair

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

type Ticket struct {
	rdb *redis.Client
}

// 并发模拟 50 位用户抢票
func (t *Ticket) buy(ctx context.Context) {
	lockKey := "fair:lock"

	lock := redislock.New(
		ctx,
		t.rdb,
		lockKey,
		redislock.WithTimeout(30*time.Second), // 锁 TTL
		redislock.WithRequestTimeout(10*time.Second), // 10s无法获取锁则放弃
	)

	var wg sync.WaitGroup
	userCount := 50
	wg.Add(userCount)

	for i := 0; i < userCount; i++ {
		go func(userId int) {
			defer wg.Done()
			requestId := fmt.Sprintf("user:%d:%s", userId, uuid.New().String())

			// 自旋公平锁 —— 在 N 秒内一直尝试
			if err := lock.SpinFairLock(requestId, 10*time.Second); err != nil {
				log.Printf("[%s] 排队超时，未抢到锁: %v", requestId, err)
				return
			}

			// 下面开始处于临界区，只有队首线程能执行
			defer lock.FairUnLock(requestId)

			// 检查剩余票数
			// 扣减库存 / 为该用户锁定票资源

			// 抢票成功
			log.Printf("[%s] 抢票成功！", requestId)
		}(i + 1)
	}

	wg.Wait()

	log.Printf("抢票结束")
}

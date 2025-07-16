package reentrant

import (
	"context"
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/redis/go-redis/v9"
	"time"
)

type Order struct {
	rdb *redis.Client
}

// CreateOrder 防止重复下单
func (o *Order) CreateOrder(ctx context.Context, userId int64, productId int64) error {
	lockKey := fmt.Sprintf("order_lock:%d:%d", userId, productId)
	lock := redislock.New(
		ctx,
		o.rdb,
		lockKey,
		redislock.WithTimeout(10*time.Second),
	)

	if err := lock.Lock(); err != nil {
		return fmt.Errorf("操作过于频繁，请稍后再试")
	}
	defer lock.UnLock()

	// 检查是否已存在订单
	// 创建新订单
	// ...

	return nil
}

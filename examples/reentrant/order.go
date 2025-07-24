package reentrant

import (
	"context"
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	"time"
)

type Order struct {
	rdb redislock.RedisInter
}

// CreateOrder 防止重复下单
func (o *Order) CreateOrder(ctx context.Context, userId int64, productId int64) error {
	lockKey := fmt.Sprintf("order_lock:%d:%d", userId, productId)
	lock := redislock.New(
		o.rdb,
		lockKey,
		redislock.WithTimeout(10*time.Second),
	)

	if err := lock.Lock(ctx); err != nil {
		return fmt.Errorf("操作过于频繁，请稍后再试")
	}
	defer lock.UnLock(ctx)

	// 检查是否已存在订单
	// 创建新订单
	// ...

	return nil
}

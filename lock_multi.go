package go_redislock

import (
	"context"
	"fmt"
	"time"
)

func (l *RedisLock) MultiLock(ctx context.Context, locks []RedisLockInter, timeout time.Duration) error {
	if len(locks) == 0 {
		return fmt.Errorf("no locks provided")
	}

	deadline := time.Now().Add(timeout)
	locked := make([]RedisLockInter, 0, len(locks))

	for _, lock := range locks {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			// 超时，回滚已加锁
			_ = l.MultiUnLock(ctx, locked)
			return fmt.Errorf("timeout trying to acquire all locks")
		}

		err := lock.Lock(ctx)
		if err != nil {
			// 当前锁失败，回滚已加锁
			_ = l.MultiUnLock(ctx, locked)
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		locked = append(locked, lock)
	}

	return nil // 全部加锁成功
}

func (l *RedisLock) MultiUnLock(ctx context.Context, locks []RedisLockInter) error {
	var lastErr error
	for _, lock := range locks {
		if err := lock.UnLock(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"time"
)

var (
	//go:embed lua/reentrantLock.lua
	reentrantLockScript string
	//go:embed lua/reentrantUnLock.lua
	reentrantUnLockScript string
	//go:embed lua/reentrantRenew.lua
	reentrantRenewScript string
)

// Lock 可重入锁加锁
func (l *RedisLock) Lock(ctx context.Context) error {
	result, err := l.redis.Eval(ctx, reentrantLockScript,
		[]string{l.key},
		l.token,
		l.lockTimeout.Milliseconds(),
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}
	if result != 1 {
		return ErrLockFailed
	}

	if l.isAutoRenew {
		ctxRenew, cancel := context.WithCancel(ctx)
		l.autoRenewCancel = cancel
		go l.autoRenew(ctxRenew)
	}

	return nil
}

// UnLock 解锁
func (l *RedisLock) UnLock(ctx context.Context) error {
	// 如果已经创建了取消函数，则执行取消操作
	if l.autoRenewCancel != nil {
		l.autoRenewCancel()
	}

	result, err := l.redis.Eval(
		ctx,
		reentrantUnLockScript,
		[]string{l.key}, l.token,
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}
	if result != 1 {
		return ErrUnLockFailed
	}

	return nil
}

// SpinLock 自旋锁
func (l *RedisLock) SpinLock(ctx context.Context, timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		if time.Now().After(exp) {
			return ErrSpinLockTimeout
		}

		// 加锁成功直接返回
		if err := l.Lock(ctx); err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-ctx.Done():
			return errors.Join(ErrSpinLockDone, context.Canceled) // 处理取消操作
		case <-time.After(100 * time.Millisecond):
			// 继续尝试下一轮加锁
		}
	}
}

// Renew 锁手动续期
func (l *RedisLock) Renew(ctx context.Context) error {
	res, err := l.redis.Eval(
		ctx,
		reentrantRenewScript,
		[]string{l.key},
		l.token,
		l.lockTimeout.Milliseconds(),
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	if res != 1 {
		return ErrLockRenewFailed
	}

	return nil
}

// 锁自动续期
func (l *RedisLock) autoRenew(ctx context.Context) {
	ticker := time.NewTicker(l.lockTimeout / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := l.Renew(ctx)
			if err != nil {
				log.Printf("Error: autoRenew failed, %v", err)
				return
			}
		}
	}
}

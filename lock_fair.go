package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"time"
)

var (
	//go:embed lua/fairLock.lua
	fairLockScript string
	//go:embed lua/fairUnlock.lua
	fairUnLockScript string
	//go:embed lua/fairRenew.lua
	fairRenewScript string
)

// FairLock 公平锁尝试加锁
// 公平锁确保请求按照顺序获取锁，避免饥饿现象
// 如果是队首且成功获取锁则返回 nil，否则返回 ErrLockFailed
func (l *RedisLock) FairLock(ctx context.Context, requestId string) error {
	result, err := l.redis.Eval(ctx, fairLockScript,
		[]string{l.key},
		requestId,
		l.lockTimeout.Milliseconds(),
		l.requestTimeout.Milliseconds(),
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	// 没有抢到锁，则进入排队，不是ok则说明不是队首
	if result != 1 {
		return ErrLockFailed
	}

	if l.isAutoRenew {
		ctxRenew, cancel := context.WithCancel(ctx)
		l.autoRenewCancel = cancel
		go l.autoFairRenew(ctxRenew, requestId)
	}

	return nil
}

// SpinFairLock 自旋公平锁
func (l *RedisLock) SpinFairLock(ctx context.Context, requestId string, timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		// 检查自旋锁是否超时
		if time.Now().After(exp) {
			return ErrSpinLockTimeout
		}

		// 尝试公平锁锁成功
		if err := l.FairLock(ctx, requestId); err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-ctx.Done(): // 检查上下文是否已取消
			return errors.Join(ErrSpinLockDone, context.Canceled)
		case <-time.After(100 * time.Millisecond):
			// 继续尝试下一轮加锁
		}
	}
}

// FairUnLock 公平锁解锁
func (l *RedisLock) FairUnLock(ctx context.Context, requestId string) error {
	if l.autoRenewCancel != nil {
		l.autoRenewCancel()
	}

	result, err := l.redis.Eval(
		ctx,
		fairUnLockScript,
		[]string{l.key},
		requestId,
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	if result != 1 {
		return ErrUnLockFailed
	}

	return nil
}

// FairRenew 公平锁手动续期
func (l *RedisLock) FairRenew(ctx context.Context, requestId string) error {
	res, err := l.redis.Eval(
		ctx,
		fairRenewScript,
		[]string{l.key},
		requestId,
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
func (l *RedisLock) autoFairRenew(ctx context.Context, requestId string) {
	ticker := time.NewTicker(l.lockTimeout / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := l.FairRenew(ctx, requestId)
			if err != nil {
				log.Printf("Error: autoFairRenew failed, Err: %v \n", err)
				return
			}
		}
	}
}

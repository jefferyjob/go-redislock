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

// Lock tries to acquire a standard lock.
// This implementation supports "reentrant locks". If the lock is currently held by the same key+token, reentry is allowed and the count is increased. Unlock() needs to be called a corresponding number of times to release the lock.
//
// Lock 尝试获取普通锁。
// 该实现支持“可重入锁”，如果当前已由相同 key+token 持有，允许重入并增加计数。需调用相应次数 Unlock() 释放
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

// UnLock releases the standard lock.
// If it is a reentrant lock, each call will reduce the holding count until the count reaches 0 and the lock will be released.
//
// UnLock 释放普通锁。
// 如果为重入锁，每调用一次减少一次持有计数，直到计数为 0 锁会被释放
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

// SpinLock keeps trying to acquire the lock until timeout.
// SpinLock 在指定超时时间内不断尝试加锁。
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

// Renew manually extends the lock expiration.
// Renew 手动延长锁的有效期。
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

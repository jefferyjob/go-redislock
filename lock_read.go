package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"time"
)

var (
	//go:embed lua/readLock.lua
	readLockScript string
	//go:embed lua/readUnLock.lua
	readUnLockScript string
	//go:embed lua/readRenew.lua
	readRenewScript string
)

func (l *RedisLock) RLock(ctx context.Context) error {
	res, err := l.redis.Eval(ctx, readLockScript,
		[]string{l.key},
		l.token,
		l.lockTimeout.Milliseconds(),
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	if res != 1 {
		return ErrLockFailed
	}

	if l.isAutoRenew {
		ctxRenew, cancel := context.WithCancel(ctx)
		l.autoRenewCancel = cancel
		go l.autoRLockRenew(ctxRenew)
	}

	return nil
}

func (l *RedisLock) RUnLock(ctx context.Context) error {
	// 如果已经创建了取消函数，则执行取消操作
	if l.autoRenewCancel != nil {
		l.autoRenewCancel()
	}

	res, err := l.redis.Eval(
		ctx,
		readUnLockScript,
		[]string{l.key}, l.token,
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}
	if res != 1 {
		return ErrUnLockFailed
	}

	return nil
}

func (l *RedisLock) SpinRLock(ctx context.Context, timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		if time.Now().After(exp) {
			return ErrSpinLockTimeout
		}

		// 加锁成功直接返回
		if err := l.RLock(ctx); err == nil {
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

func (l *RedisLock) RRenew(ctx context.Context) error {
	res, err := l.redis.Eval(
		ctx,
		readRenewScript,
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
func (l *RedisLock) autoRLockRenew(ctx context.Context) {
	ticker := time.NewTicker(l.lockTimeout / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := l.RRenew(ctx)
			if err != nil {
				log.Printf("Error: autoRRenew failed, %v", err)
				return
			}
		}
	}
}

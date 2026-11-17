package go_redislock

import (
	"context"
	_ "embed"
	"errors"
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

	// 可重入加锁时复用已在运行的续期 goroutine，避免泄漏
	l.startAutoRenew(l.RRenew)

	return nil
}

func (l *RedisLock) RUnLock(ctx context.Context) error {
	res, err := l.redis.Eval(
		ctx,
		readUnLockScript,
		[]string{l.key}, l.token,
	).Int64()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	switch res {
	case 2:
		// 部分解锁：本持有者仍持有读锁，自动续期继续保持
		return nil
	case 1:
		// 本持有者的读锁已完全释放：停止自动续期
		l.stopAutoRenew()
		return nil
	default:
		return ErrUnLockFailed
	}
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

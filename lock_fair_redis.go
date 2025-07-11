package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"time"
)

var (
	//go:embed lua/fairLock.lua
	fairLockScript string
	//go:embed lua/fairUnlock.lua
	fairUnLockScript string
)

// FairLock 公平锁尝试加锁
// 公平锁确保请求按照顺序获取锁，避免饥饿现象
// 如果是队首且成功获取锁则返回 nil，否则返回 ErrLockFailed
func (l *RedisLock) FairLock(requestId string) error {
	result, err := l.redis.Eval(l.Context, fairLockScript,
		[]string{l.key},
		requestId,
		l.lockTimeout.Seconds(),
		l.requestTimeout.Seconds(),
	).Int()

	if err != nil {
		return errors.Join(err, ErrException)
	}

	// 没有抢到锁，则进入排队，不是ok则说明不是队首
	if result != 1 {
		return ErrLockFailed
	}

	if l.isAutoRenew {
		l.autoRenewCtx, l.autoRenewCancel = context.WithCancel(l.Context)
		go l.autoRenew()
	}

	return nil
}

// SpinFairLock 自旋公平锁
func (l *RedisLock) SpinFairLock(requestId string, timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		// 检查自旋锁是否超时
		if time.Now().After(exp) {
			return ErrSpinLockTimeout
		}

		// 尝试公平锁锁成功
		err := l.FairLock(requestId)
		if err == nil {
			return nil
		}

		// 如果是 ErrLockFailed，说明没有抢到锁，继续自旋
		if errors.Is(err, ErrLockFailed) {
			select {
			case <-l.Context.Done(): // 检查上下文是否已取消
				return ErrSpinLockDone
			default:
				time.Sleep(100 * time.Millisecond) // 等待一段时间后重试
			}
		} else {
			return err // 其他错误直接返回
		}
	}
}

// FairUnLock 公平锁解锁
func (l *RedisLock) FairUnLock(requestId string) error {
	if l.autoRenewCancel != nil {
		l.autoRenewCancel()
	}

	result, err := l.redis.Eval(
		l.Context,
		fairUnLockScript,
		[]string{l.key},
		requestId,
	).Int()

	if err != nil {
		return ErrException
	}

	if result != 1 {
		return ErrUnLockFailed
	}

	return nil
}

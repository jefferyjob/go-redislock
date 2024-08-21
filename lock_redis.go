package go_redislock

import (
	"context"
	_ "embed"
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

// Lock 加锁
func (lock *RedisLock) Lock() error {
	result, err := lock.redis.Eval(lock.Context, reentrantLockScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrLockFailed
	}

	if lock.isAutoRenew {
		lock.autoRenewCtx, lock.autoRenewCancel = context.WithCancel(lock.Context)
		go lock.autoRenew()
	}

	return nil
}

// UnLock 解锁
func (lock *RedisLock) UnLock() error {
	// 如果已经创建了取消函数，则执行取消操作
	if lock.autoRenewCancel != nil {
		lock.autoRenewCancel()
	}

	// 通知当前已解锁
	defer func() {
		lock.unLockChan = make(chan struct{})
	}()

	result, err := lock.redis.Eval(lock.Context, reentrantUnLockScript, []string{lock.key}, lock.token).Result()

	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrUnLockFailed
	}
	return nil
}

// SpinLock 自旋锁
func (lock *RedisLock) SpinLock(timeout time.Duration) error {
	timeDeadline := time.After(timeout)
	for {
		// 加锁成功直接返回
		if err := lock.Lock(); err == nil {
			return nil
		}

		// 如果加锁失败，则等待下次可加锁时间
		select {
		case <-lock.Context.Done():
			return ErrSpinLockDone // 处理取消操作
		case <-lock.unLockChan:
			// 继续下一次尝试
		case <-timeDeadline:
			return ErrSpinLockTimeout
		}
	}
}

// Renew 锁手动续期
func (lock *RedisLock) Renew() error {
	res, err := lock.redis.Eval(lock.Context, reentrantRenewScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}
	if res != "OK" {
		return ErrLockRenewFailed
	}

	return nil
}

// 锁自动续期
func (lock *RedisLock) autoRenew() {
	ticker := time.NewTicker(lock.lockTimeout / 3)
	defer ticker.Stop()

	for {
		select {
		case <-lock.autoRenewCtx.Done():
			return
		case <-ticker.C:
			err := lock.Renew()
			if err != nil {
				log.Printf("Error: autoRenew failed, %v", err)
				return
			}
		}
	}
}

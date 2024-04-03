package go_redislock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// Lock 加锁
func (lock *RedisLock) Lock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.RedisClientInter.Eval(lock.Context, lockScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result != "OK" {
		return errors.New("lock acquisition failed")
	}

	if lock.isAutoRenew {
		lock.autoRenewCtx, lock.autoRenewCancel = context.WithCancel(lock.Context)
		go lock.autoRenew()
	}

	return nil
}

// UnLock 解锁
func (lock *RedisLock) UnLock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	// 如果已经创建了取消函数，则执行取消操作
	if lock.autoRenewCancel != nil {
		lock.autoRenewCancel()
	}

	result, err := lock.RedisClientInter.Eval(lock.Context, unLockScript, []string{lock.key}, lock.token).Result()

	if err != nil {
		return fmt.Errorf("ailed to release lock: %w", err)
	}

	if result != "OK" {
		return errors.New("lock release failed")
	}

	return nil
}

// SpinLock 自旋锁
func (lock *RedisLock) SpinLock(timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		if time.Now().After(exp) {
			return errors.New("spin lock timeout")
		}

		// 加锁成功直接返回
		err := lock.Lock()
		if err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-lock.Context.Done():
			return lock.Context.Err() // 处理取消操作
		case <-time.After(100 * time.Millisecond):
			// 继续尝试下一轮加锁
		}
	}
}

// Renew 锁手动续期
func (lock *RedisLock) Renew() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	res, err := lock.RedisClientInter.Eval(lock.Context, renewScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return fmt.Errorf("failed to renew lock: %s", err)
	}

	if res != "OK" {
		return errors.New("lock renewal failed")
	}

	return nil
}

// 锁自动续期
func (lock *RedisLock) autoRenew() {
	ticker := time.NewTicker(lock.lockTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-lock.autoRenewCtx.Done():
			return
		case <-ticker.C:
			err := lock.Renew()
			if err != nil {
				log.Println("autoRenew failed:", err)
				return
			}
		}
	}
}

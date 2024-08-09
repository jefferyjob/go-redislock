package go_redislock

import (
	"context"
	_ "embed"
	"log"
	"time"
)

var (
	//go:embed lua/reentrantLock.lua
	lockScript string
	//go:embed lua/reentrantUnLock.lua
	unLockScript string
	//go:embed lua/reentrantRenew.lua
	renewScript string
)

// Lock 加锁
func (lock *RedisLock) Lock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, lockScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

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
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	// 如果已经创建了取消函数，则执行取消操作
	if lock.autoRenewCancel != nil {
		lock.autoRenewCancel()
	}

	result, err := lock.redis.Eval(lock.Context, unLockScript, []string{lock.key}, lock.token).Result()

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
	exp := time.Now().Add(timeout)
	for {
		if time.Now().After(exp) {
			return ErrSpinLockTimeout
		}

		// 加锁成功直接返回
		if err := lock.Lock(); err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-lock.Context.Done():
			return ErrSpinLockDone // 处理取消操作
		case <-time.After(100 * time.Millisecond):
			// 继续尝试下一轮加锁
		}
	}
}

// Renew 锁手动续期
func (lock *RedisLock) Renew() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	res, err := lock.redis.Eval(lock.Context, renewScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

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

// FairLock 公平锁加锁
func (lock *RedisLock) FairLock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "fairLock.lua", []string{lock.key}, lock.token, lock.lockTimeout.Seconds(), lockQueueTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockFailed
	}

	return nil
}

// FairUnLock 公平锁解锁
func (lock *RedisLock) FairUnLock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "fairUnlock.lua", []string{lock.key}, lock.token).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrUnLockFailed
	}

	return nil
}

// FairRenew 公平锁续期
func (lock *RedisLock) FairRenew() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	res, err := lock.redis.Eval(lock.Context, "fairRenew.lua", []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if res != "OK" {
		return ErrLockRenewFailed
	}

	return nil
}

// MultiLock 联锁加锁
func (lock *RedisLock) MultiLock(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "acquireLocks.lua", lockKeys, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockFailed
	}

	return nil
}

// MultiUnLock 联锁解锁
func (lock *RedisLock) MultiUnLock(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "releaseLocks.lua", lockKeys, lock.token).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrUnLockFailed
	}

	return nil
}

// MultiRenew 联锁续期
func (lock *RedisLock) MultiRenew(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "renewLocks.lua", lockKeys, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockRenewFailed
	}

	return nil
}

// RedLock 红锁加锁
func (lock *RedisLock) RedLock(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "acquireRedLocks.lua", lockKeys, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockFailed
	}

	return nil
}

// RedUnLock 红锁解锁
func (lock *RedisLock) RedUnLock(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "releaseRedLocks.lua", lockKeys, lock.token).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockFailed
	}

	return nil
}

// RedRenew 红锁续期
func (lock *RedisLock) RedRenew(lockKeys []string) error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	result, err := lock.redis.Eval(lock.Context, "renewRedLocks.lua", lockKeys, lock.token, lock.lockTimeout.Seconds()).Result()

	if err != nil {
		return ErrException
	}

	if result != "OK" {
		return ErrLockRenewFailed
	}

	return nil
}

联锁的逻辑旨在确保一组分布式锁中的所有锁都能成功获取，才能确保整体操作的原子性。简单来说，联锁保证了在多个锁中至少一个锁失败时，所有锁都会被释放，以避免部分成功、部分失败的情况。

### 联锁逻辑的实现

#### 1. 联锁逻辑

联锁的关键在于：

1. **获取所有锁**：确保所有指定的锁都被成功获取。
2. **释放所有锁**：如果获取锁的过程失败，需要释放已成功获取的锁。

#### 2. 联锁的 Lua 脚本

我们需要两个 Lua 脚本：

1. **获取锁的 Lua 脚本**：用于在 Redis 中获取一组锁。
2. **释放锁的 Lua 脚本**：用于释放一组锁。

**获取锁的 Lua 脚本 (`acquireLocks.lua`)**：

```lua
local lock_keys = KEYS
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local acquired = {}

-- 尝试获取所有锁
for i, lock_key in ipairs(lock_keys) do
    local result = redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl)
    if result == 'OK' then
        table.insert(acquired, lock_key)
    else
        -- 如果获取锁失败，则释放所有已成功获取的锁
        for _, key in ipairs(acquired) do
            redis.call('DEL', key)
        end
        return nil
    end
end

return 'OK'
```

**释放锁的 Lua 脚本 (`releaseLocks.lua`)**：

```lua
local lock_keys = KEYS
local request_id = ARGV[1]

-- 尝试释放所有锁
for _, lock_key in ipairs(lock_keys) do
    if redis.call('GET', lock_key) == request_id then
        redis.call('DEL', lock_key)
    end
end

return 'OK'
```

### 更新的 Go 代码

我们需要在 Go 代码中添加方法来使用这些 Lua 脚本。

```go
package go_redislock

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

// 默认锁超时时间
const lockTime = 5 * time.Second

type RedisLockInter interface {
	Lock() error
	UnLock() error
	SpinLock(timeout time.Duration) error
	Renew() error
	FairLock() error
	FairUnLock() error
	FairRenew() error
	AcquireLocks(lockKeys []string) error
	ReleaseLocks(lockKeys []string) error
}

type RedisInter interface {
	redis.Scripter
}

type RedisLock struct {
	context.Context
	redis           RedisInter
	key             string
	token           string
	lockTimeout     time.Duration
	isAutoRenew     bool
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
	mutex           sync.Mutex
}

type Option func(lock *RedisLock)

func New(ctx context.Context, redisClient RedisInter, lockKey string, options ...Option) RedisLockInter {
	lock := &RedisLock{
		Context:     ctx,
		redis:       redisClient,
		lockTimeout: lockTime,
	}
	for _, f := range options {
		f(lock)
	}

	lock.key = lockKey

	if lock.token == "" {
		lock.token = fmt.Sprintf("lock_token:%s", uuid.New().String())
	}

	return lock
}

// WithTimeout 设置锁过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *RedisLock) {
		lock.isAutoRenew = true
	}
}

// WithToken 设置锁的Token
func WithToken(token string) Option {
	return func(lock *RedisLock) {
		lock.token = token
	}
}

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

		if err := lock.Lock(); err == nil {
			return nil
		}

		select {
		case <-lock.Context.Done():
			return ErrSpinLockDone
		case <-time.After(100 * time.Millisecond):
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

// AcquireLocks 联锁加锁
func (lock *RedisLock) AcquireLocks(lockKeys []string) error {
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

// ReleaseLocks 联锁解锁
func (lock *RedisLock) ReleaseLocks(lockKeys []string) error {
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
```

### 关键点解释

1. **获取锁 (`acquireLocks.lua`)**

    - 依次尝试获取所有指定的锁。如果某个锁获取失败，则释放所有已成功获取的锁。
    - 使用 `ZADD` 将请求的时间戳和请求 ID 添加到有序集合中，保持锁的公平性。

2. **释放锁 (`releaseLocks.lua`)**

    - 释放所有指定的锁。如果锁的持有者 ID 匹配当前请求 ID，则删除该锁。

3. **Go 代码更新**

    - 添加 `AcquireLocks` 和 `ReleaseLocks` 方法，用于执行获取和释放多个锁的操作。

这样实现的联锁可以确保在获取多个锁时，所有锁都成功获取才认为操作成功。如果获取过程中
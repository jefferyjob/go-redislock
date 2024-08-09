为了完整地实现一个分布式读写锁，包含加锁、解锁和续期功能，下面是详细的实现，包括 Lua 脚本和 Go 代码。

### 1. Lua 脚本实现

#### 1.1. 读锁 Lua 脚本

**功能**：获取读锁。

```lua
-- 获取读锁 Lua 脚本
local lock_key = KEYS[1]
local read_count_key = lock_key .. ':read_count'
local write_lock_key = lock_key .. ':write_lock'
local request_id = ARGV[1]
local request_timeout = tonumber(ARGV[2])

-- 检查写锁是否存在
local write_lock_exists = redis.call('EXISTS', write_lock_key)
if write_lock_exists == 1 then
    return nil
end

-- 获取当前时间
local current_time = tonumber(redis.call('TIME')[1])

-- 添加读锁
local read_count = tonumber(redis.call('GET', read_count_key) or '0')
redis.call('SET', read_count_key, read_count + 1)

-- 设置读锁过期时间
redis.call('EXPIRE', read_count_key, request_timeout)

return 'OK'
```

#### 1.2. 写锁 Lua 脚本

**功能**：获取写锁。

```lua
-- 获取写锁 Lua 脚本
local lock_key = KEYS[1]
local write_lock_key = lock_key .. ':write_lock'
local request_id = ARGV[1]
local lock_timeout = tonumber(ARGV[2])

-- 检查是否有读锁
local read_count_key = lock_key .. ':read_count'
local read_count = tonumber(redis.call('GET', read_count_key) or '0')
if read_count > 0 then
    return nil
end

-- 尝试获取写锁
local result = redis.call('SET', write_lock_key, request_id, 'NX', 'EX', lock_timeout)
if result == 'OK' then
    return 'OK'
end

return nil
```

#### 1.3. 释放读锁 Lua 脚本

**功能**：释放读锁。

```lua
-- 释放读锁 Lua 脚本
local lock_key = KEYS[1]
local read_count_key = lock_key .. ':read_count'
local request_id = ARGV[1]

-- 获取当前读锁计数
local read_count = tonumber(redis.call('GET', read_count_key) or '0')

if read_count > 0 then
    -- 释放一个读锁
    redis.call('SET', read_count_key, read_count - 1)
    
    -- 如果读锁计数为0，检查写锁
    if read_count - 1 == 0 then
        redis.call('DEL', lock_key .. ':write_lock')
    end
end

return 'OK'
```

#### 1.4. 释放写锁 Lua 脚本

**功能**：释放写锁。

```lua
-- 释放写锁 Lua 脚本
local lock_key = KEYS[1]
local write_lock_key = lock_key .. ':write_lock'
local request_id = ARGV[1]

-- 检查写锁是否存在
local current_lock_value = redis.call('GET', write_lock_key)
if current_lock_value == request_id then
    redis.call('DEL', write_lock_key)
    return 'OK'
end

return nil
```

#### 1.5. 续期写锁 Lua 脚本

**功能**：续期写锁。

```lua
-- 续期写锁 Lua 脚本
local lock_key = KEYS[1]
local write_lock_key = lock_key .. ':write_lock'
local request_id = ARGV[1]
local lock_timeout = tonumber(ARGV[2])

-- 检查写锁是否存在并且是当前请求持有的
local current_lock_value = redis.call('GET', write_lock_key)
if current_lock_value == request_id then
    redis.call('EXPIRE', write_lock_key, lock_timeout)
    return 'OK'
end

return nil
```

### 2. Go 代码实现

下面是 Go 代码，用于实现分布式读写锁，包括加锁、解锁和续期功能。

```go
package go_redislock

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisLockInter interface {
	// ReadLock 获取读锁
	ReadLock(requestID string, timeout time.Duration) error
	// WriteLock 获取写锁
	WriteLock(requestID string, timeout time.Duration) error
	// ReleaseReadLock 释放读锁
	ReleaseReadLock(requestID string) error
	// ReleaseWriteLock 释放写锁
	ReleaseWriteLock(requestID string) error
	// RenewWriteLock 续期写锁
	RenewWriteLock(requestID string, timeout time.Duration) error
}

type RedisLock struct {
	context.Context
	redis     RedisInter
	lockKey   string
	lockTTL   time.Duration
	renewTTL  time.Duration
}

func NewRedisLock(ctx context.Context, redisClient RedisInter, lockKey string, lockTTL, renewTTL time.Duration) *RedisLock {
	return &RedisLock{
		Context:  ctx,
		redis:    redisClient,
		lockKey:  lockKey,
		lockTTL:  lockTTL,
		renewTTL: renewTTL,
	}
}

func (lock *RedisLock) ReadLock(requestID string, timeout time.Duration) error {
	script := `
		local lock_key = KEYS[1]
		local read_count_key = lock_key .. ':read_count'
		local write_lock_key = lock_key .. ':write_lock'
		local request_id = ARGV[1]
		local request_timeout = tonumber(ARGV[2])

		local write_lock_exists = redis.call('EXISTS', write_lock_key)
		if write_lock_exists == 1 then
			return nil
		end

		local current_time = tonumber(redis.call('TIME')[1])

		local read_count = tonumber(redis.call('GET', read_count_key) or '0')
		redis.call('SET', read_count_key, read_count + 1)
		redis.call('EXPIRE', read_count_key, request_timeout)

		return 'OK'
	`
	_, err := lock.redis.Eval(lock.Context, script, []string{lock.lockKey}, requestID, timeout.Seconds()).Result()
	return err
}

func (lock *RedisLock) WriteLock(requestID string, timeout time.Duration) error {
	script := `
		local lock_key = KEYS[1]
		local write_lock_key = lock_key .. ':write_lock'
		local request_id = ARGV[1]
		local lock_timeout = tonumber(ARGV[2])

		local read_count_key = lock_key .. ':read_count'
		local read_count = tonumber(redis.call('GET', read_count_key) or '0')
		if read_count > 0 then
			return nil
		end

		local result = redis.call('SET', write_lock_key, request_id, 'NX', 'EX', lock_timeout)
		if result == 'OK' then
			return 'OK'
		end

		return nil
	`
	_, err := lock.redis.Eval(lock.Context, script, []string{lock.lockKey}, requestID, timeout.Seconds()).Result()
	return err
}

func (lock *RedisLock) ReleaseReadLock(requestID string) error {
	script := `
		local lock_key = KEYS[1]
		local read_count_key = lock_key .. ':read_count'
		local request_id = ARGV[1]

		local read_count = tonumber(redis.call('GET', read_count_key) or '0')

		if read_count > 0 then
			redis.call('SET', read_count_key, read_count - 1)
			
			if read_count - 1 == 0 then
				redis.call('DEL', lock_key .. ':write_lock')
			end
		end

		return 'OK'
	`
	_, err := lock.redis.Eval(lock.Context, script, []string{lock.lockKey}, requestID).Result()
	return err
}

func (lock *RedisLock) ReleaseWriteLock(requestID string) error {
	script := `
		local lock_key = KEYS[1]
		local write_lock_key = lock_key .. ':write_lock'
		local request_id = ARGV[1]

		local current_lock_value = redis.call('GET', write_lock_key)
		if current_lock_value == request_id then
			redis.call('DEL', write_lock_key)
			return 'OK'
		end

		return nil
	`
	_, err := lock.redis.Eval(lock.Context, script, []string{lock.lockKey}, requestID).Result()
	return err
}

func (lock *RedisLock) RenewWriteLock(requestID string, timeout time.Duration) error {
	script := `
		local lock_key = KEYS[1]
		local write_lock_key = lock_key .. ':write_lock'
		local request_id = ARGV[1]
		local lock_timeout = tonumber(ARGV[2])

		local current_lock_value = redis.call('GET', write_lock_key)
		if current_lock_value == request_id then
			redis.call('EXPIRE', write_lock_key, lock_timeout)
			return 'OK'
		end

		return nil
	`
	_, err := lock.redis.Eval(lock.Context, script, []string{lock.lockKey}, requestID, timeout.Seconds()).Result()
	return err
}
```

### 关键点总结



1. **读锁**：
    - 允许多个线程同时读取资源。
    - 在持有读锁时，不允许获得写锁。
    - 通过 Lua 脚本管理读锁计数和超时设置。

2. **写锁**：
    - 允许一个线程写入资源。
    - 在持有写锁时，不允许其他线程获得读锁或写锁。
    - 通过 Lua 脚本检查是否存在读锁，并尝试获取写锁。

3. **释放锁**：
    - 释放读锁时减少读锁计数，并在计数为零时删除写锁。
    - 释放写锁时删除写锁，允许其他线程获取读锁或写锁。

4. **续期锁**：
    - 续期写锁时延长写锁的过期时间，以确保锁在使用期间不会过期。

这些实现能够在 Redis 中有效地管理读写锁的状态，确保分布式系统中的并发访问控制。
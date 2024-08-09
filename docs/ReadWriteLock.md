分布式读写锁是分布式系统中的一种锁机制，用于在多个实例之间协调对共享资源的访问。读写锁的设计目的是在允许多个读操作并发执行的同时，确保写操作是互斥的。即，多个读操作可以同时进行，但写操作必须是排他的，不能与其他读或写操作同时进行。

### 读写锁的基本概念

1. **读操作**（Read Lock）：
    - 允许多个线程或进程并发地进行读操作。
    - 只要没有写操作在进行，多个读操作可以同时执行。

2. **写操作**（Write Lock）：
    - 写操作是排他的，意味着在进行写操作时，不允许有其他读或写操作进行。
    - 写锁请求必须等待所有正在进行的读操作和写操作完成后才能获得。

### 分布式读写锁的特点

- **读锁**：当一个节点请求读锁时，系统可以允许多个节点同时获得读锁，只要没有节点持有写锁。
- **写锁**：当一个节点请求写锁时，系统会阻止所有其他节点（包括读锁持有者）访问被锁定的资源，直到写锁释放。

### 分布式读写锁的实现

实现分布式读写锁通常包括以下几个步骤：

1. **获取读锁**：
    - 节点在多个 Redis 实例上设置一个标识符来表示读锁的存在。
    - 只要写锁没有被持有，节点可以并发地获取读锁。

2. **获取写锁**：
    - 节点在多个 Redis 实例上尝试设置写锁，确保所有节点在设置写锁时能够看到没有其他节点持有写锁或读锁。
    - 一旦获取到写锁，所有其他节点（包括读锁持有者）都不能再获取读锁或写锁，直到写锁释放。

3. **释放锁**：
    - 节点释放读锁时，其他节点可以继续获取读锁，但写锁请求可以继续排队。
    - 节点释放写锁时，所有等待中的读锁和写锁请求可以被处理。

### 分布式读写锁的 Lua 脚本示例

以下是一个简化的分布式读写锁的 Lua 脚本示例。它假设有一个读锁和一个写锁的 Redis 键。

#### 1. 获取读锁的 Lua 脚本 (`readlock.lua`)

```lua
local lock_key = KEYS[1]
local read_lock_key = lock_key .. ':read'
local write_lock_key = lock_key .. ':write'
local request_id = ARGV[1]

-- 检查是否有写锁
if redis.call('EXISTS', write_lock_key) == 1 then
    return nil
end

-- 增加读锁计数
redis.call('INCR', read_lock_key)
return "OK"
```

#### 2. 释放读锁的 Lua 脚本 (`release_readlock.lua`)

```lua
local lock_key = KEYS[1]
local read_lock_key = lock_key .. ':read'
local request_id = ARGV[1]

-- 减少读锁计数
local count = tonumber(redis.call('GET', read_lock_key) or '0')
if count > 1 then
    redis.call('DECR', read_lock_key)
else
    redis.call('DEL', read_lock_key)
end

return "OK"
```

#### 3. 获取写锁的 Lua 脚本 (`writelock.lua`)

```lua
local lock_key = KEYS[1]
local write_lock_key = lock_key .. ':write'
local request_id = ARGV[1]

-- 尝试获取写锁
if redis.call('SETNX', write_lock_key, request_id) == 1 then
    redis.call('EXPIRE', write_lock_key, tonumber(ARGV[2]))
    return "OK"
end

return nil
```

#### 4. 释放写锁的 Lua 脚本 (`release_writelock.lua`)

```lua
local lock_key = KEYS[1]
local write_lock_key = lock_key .. ':write'
local request_id = ARGV[1]

-- 释放写锁
if redis.call('GET', write_lock_key) == request_id then
    redis.call('DEL', write_lock_key)
    return "OK"
end

return nil
```

### Go 代码实现

在 Go 代码中，我们需要实现对这些 Lua 脚本的调用，以获取和释放读锁及写锁。

```go
package go_redislock

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// ReadLock 获取读锁
func (lock *RedisLock) ReadLock() error {
	result, err := lock.redis.Eval(lock.Context, "readlock.lua", []string{lock.key}, lock.token).Result()
	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrReadLockFailed
	}
	return nil
}

// ReleaseReadLock 释放读锁
func (lock *RedisLock) ReleaseReadLock() error {
	result, err := lock.redis.Eval(lock.Context, "release_readlock.lua", []string{lock.key}, lock.token).Result()
	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrReleaseReadLockFailed
	}
	return nil
}

// WriteLock 获取写锁
func (lock *RedisLock) WriteLock() error {
	result, err := lock.redis.Eval(lock.Context, "writelock.lua", []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()
	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrWriteLockFailed
	}
	return nil
}

// ReleaseWriteLock 释放写锁
func (lock *RedisLock) ReleaseWriteLock() error {
	result, err := lock.redis.Eval(lock.Context, "release_writelock.lua", []string{lock.key}, lock.token).Result()
	if err != nil {
		return ErrException
	}
	if result != "OK" {
		return ErrReleaseWriteLockFailed
	}
	return nil
}
```

### 关键点解释

1. **获取读锁**：
    - 检查是否有写锁存在。如果有，则无法获取读锁。
    - 如果没有写锁，则增加读锁计数。

2. **释放读锁**：
    - 减少读锁计数。如果计数减到 0，则删除读锁键。

3. **获取写锁**：
    - 尝试设置写锁。如果成功，则设置过期时间。

4. **释放写锁**：
    - 如果当前请求持有写锁，则删除写锁键。

通过这些实现，你可以在分布式系统中使用读写锁来协调对共享资源的访问，优化读操作的并发性能，并保证写操作的排他性。
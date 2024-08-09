红锁（Redlock）是一种分布式锁算法，旨在确保在多个 Redis 实例上实现高可用和高容错的分布式锁。其核心思想是通过在多个 Redis 实例上获取锁，并要求大多数实例成功获取锁，来保证锁的有效性和一致性。

### 红锁的实现

红锁的核心步骤包括：

1. **获取锁**：在多个 Redis 实例上同时尝试获取锁，要求大多数实例成功获取锁。
2. **释放锁**：在多个 Redis 实例上释放锁，要求大多数实例成功释放锁。
3. **续期锁**：在多个 Redis 实例上续期锁，要求大多数实例成功续期锁。

### Lua 脚本

红锁的 Lua 脚本用于在 Redis 实例上获取、释放和续期锁。

#### 1. 加锁 Lua 脚本 (`redlock_acquire.lua`)

```lua
local lock_key = KEYS[1]
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

if redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl) then
    return "OK"
end

return nil
```

#### 2. 解锁 Lua 脚本 (`redlock_release.lua`)

```lua
local lock_key = KEYS[1]
local request_id = ARGV[1]

if redis.call('GET', lock_key) == request_id then
    redis.call('DEL', lock_key)
    return "OK"
end

return nil
```

#### 3. 续期 Lua 脚本 (`redlock_renew.lua`)

```lua
local lock_key = KEYS[1]
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

if redis.call('GET', lock_key) == request_id then
    redis.call('EXPIRE', lock_key, lock_ttl)
    return "OK"
end

return nil
```

### Go 代码

在 Go 代码中，我们需要实现红锁的加锁、解锁和续期逻辑，确保在多个 Redis 实例上执行这些操作，并检查大多数实例的成功状态。

```go
package go_redislock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

const (
	redlockQuorum = 2 // 大多数实例的数量
)

// Redlock 实例表示一个 Redis 实例
type Redlock struct {
	redisClients []RedisInter
	key          string
	token        string
	lockTimeout  time.Duration
}

// NewRedlock 创建一个 Redlock 实例
func NewRedlock(redisClients []RedisInter, lockKey string, options ...Option) *Redlock {
	redlock := &Redlock{
		redisClients: redisClients,
		key:          lockKey,
		lockTimeout:  lockTime,
	}
	for _, f := range options {
		f(redlock)
	}

	if redlock.token == "" {
		redlock.token = fmt.Sprintf("lock_token:%s", uuid.New().String())
	}

	return redlock
}

// Acquire 锁加锁
func (redlock *Redlock) Acquire() error {
	start := time.Now()
	quorum := 0

	for _, redisClient := range redlock.redisClients {
		result, err := redisClient.Eval(context.Background(), "redlock_acquire.lua", []string{redlock.key}, redlock.token, redlock.lockTimeout.Seconds()).Result()
		if err != nil {
			return ErrException
		}

		if result == "OK" {
			quorum++
		}

		if time.Since(start) > redlock.lockTimeout {
			break
		}
	}

	if quorum >= redlockQuorum {
		return nil
	}

	// 如果未能达到大多数实例，则释放所有锁
	redlock.Release()
	return ErrLockFailed
}

// Release 锁解锁
func (redlock *Redlock) Release() error {
	for _, redisClient := range redlock.redisClients {
		result, err := redisClient.Eval(context.Background(), "redlock_release.lua", []string{redlock.key}, redlock.token).Result()
		if err != nil {
			return ErrException
		}

		if result != "OK" {
			return ErrUnLockFailed
		}
	}

	return nil
}

// Renew 锁续期
func (redlock *Redlock) Renew() error {
	for _, redisClient := range redlock.redisClients {
		result, err := redisClient.Eval(context.Background(), "redlock_renew.lua", []string{redlock.key}, redlock.token, redlock.lockTimeout.Seconds()).Result()
		if err != nil {
			return ErrException
		}

		if result != "OK" {
			return ErrLockRenewFailed
		}
	}

	return nil
}
```

### 关键点解释

1. **加锁 (`redlock_acquire.lua`)**：
    - 在 Redis 中尝试设置锁键，如果成功，则返回 `OK`。

2. **解锁 (`redlock_release.lua`)**：
    - 如果锁的持有者 ID 匹配当前请求 ID，则删除锁键。

3. **续期 (`redlock_renew.lua`)**：
    - 如果锁的持有者 ID 匹配当前请求 ID，则续期锁的过期时间。

4. **Go 代码中的实现**：
    - `Acquire` 方法在多个 Redis 实例上尝试获取锁，并确保大多数实例成功获取锁。如果无法达到大多数实例，则释放所有锁。
    - `Release` 方法在所有 Redis 实例上释放锁。
    - `Renew` 方法在所有 Redis 实例上续期锁。

通过实现上述逻辑，红锁能够提供高可用、高容错的分布式锁功能，确保锁在多个 Redis 实例中保持一致。
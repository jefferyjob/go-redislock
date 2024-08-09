分布式闭锁（CountDownLatch）是一种用于同步的机制，允许一个或多个线程等待直到一组操作完成。它可以确保在继续执行之前，所有的操作都已经完成。

在分布式系统中，闭锁可以通过分布式数据库（如 Redis）来实现。以下是如何使用 Redis 实现分布式闭锁的基本思路和步骤：

### 实现思路

1. **初始化闭锁**：
    - 创建一个计数器，表示需要等待的事件数量。这个计数器在 Redis 中表示为一个键，其值为等待的事件数。

2. **等待操作**：
    - 每个线程或进程会检查计数器的值，如果值大于0，则线程会等待。线程会定期检查计数器的值以决定是否继续执行。

3. **完成操作**：
    - 每个完成的操作都会减少计数器的值。如果计数器的值减少到0，则表示所有操作都已完成，等待的线程或进程可以继续执行。

### Lua 脚本实现

以下是实现分布式闭锁的 Lua 脚本示例，包括初始化闭锁、等待操作和完成操作的实现。

#### 1. 初始化闭锁的 Lua 脚本 (`init_latch.lua`)

```lua
local latch_key = KEYS[1]
local count = tonumber(ARGV[1])

-- 设置闭锁计数器
redis.call('SET', latch_key, count)
redis.call('EXPIRE', latch_key, tonumber(ARGV[2])) -- 可选: 设置过期时间
return "OK"
```

#### 2. 释放闭锁（减少计数）的 Lua 脚本 (`countdown_latch.lua`)

```lua
local latch_key = KEYS[1]

-- 减少闭锁计数器
local count = tonumber(redis.call('DECR', latch_key))

if count <= 0 then
    redis.call('DEL', latch_key) -- 可选: 删除键
    return "0"
else
    return tostring(count)
end
```

#### 3. 等待闭锁的 Lua 脚本 (`await_latch.lua`)

```lua
local latch_key = KEYS[1]
local timeout = tonumber(ARGV[1])

-- 尝试获取闭锁计数器的值
local start_time = redis.call('TIME')[1]
while true do
    local count = tonumber(redis.call('GET', latch_key))
    if not count or count <= 0 then
        return "OK"
    end

    local current_time = redis.call('TIME')[1]
    if (current_time - start_time) > timeout then
        return "TIMEOUT"
    end

    redis.call('WAIT', 0.1)
end
```

### Go 代码实现

以下是如何在 Go 中调用这些 Lua 脚本来实现分布式闭锁的示例代码。

```go
package go_redislock

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// 初始化闭锁
func (lock *RedisLock) InitLatch(ctx context.Context, latchKey string, count int, expiration time.Duration) error {
	_, err := lock.redis.Eval(ctx, "init_latch.lua", []string{latchKey}, count, int(expiration.Seconds())).Result()
	return err
}

// 释放闭锁
func (lock *RedisLock) CountDownLatch(ctx context.Context, latchKey string) (int, error) {
	count, err := lock.redis.Eval(ctx, "countdown_latch.lua", []string{latchKey}).Result()
	if err != nil {
		return 0, err
	}
	return int(count.(int64)), nil
}

// 等待闭锁
func (lock *RedisLock) AwaitLatch(ctx context.Context, latchKey string, timeout time.Duration) error {
	timeoutSec := int(timeout.Seconds())
	start := time.Now()
	for {
		count, err := lock.redis.Get(ctx, latchKey).Result()
		if err == redis.Nil || count == "0" {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for latch to be released")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
```

### 使用示例

```go
func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	lock := New(context.Background(), rdb, "latch_key", WithTimeout(10*time.Second))

	// 初始化闭锁
	lock.InitLatch(context.Background(), "my_latch", 3, 10*time.Second)

	// 等待闭锁
	go func() {
		err := lock.AwaitLatch(context.Background(), "my_latch", 60*time.Second)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Latch released!")
		}
	}()

	// 释放闭锁
	time.Sleep(5 * time.Second)
	for i := 0; i < 3; i++ {
		_, err := lock.CountDownLatch(context.Background(), "my_latch")
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
```

### 关键点

- **初始化闭锁**：设置一个计数器，表示需要等待的操作数。
- **释放闭锁**：每个完成的操作减少计数器的值。
- **等待闭锁**：线程或进程定期检查计数器的值，直到它为0或者超时。

通过这些实现，你可以在分布式系统中使用闭锁来协调多个操作，确保所有操作完成之后再继续执行。
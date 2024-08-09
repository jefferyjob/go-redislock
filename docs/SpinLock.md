### 可重入锁的实现

结合您提供的代码，实现分布式可重入锁的思路如下：

#### 实现思路

1. **持有者和计数管理**：
    - 使用 Redis 键来跟踪锁的持有者（如请求 ID）和持有者的计数（表示同一请求获取锁的次数）。

2. **获取锁**：
    - 如果锁没有被持有或持有者是当前请求，则成功获取锁，并设置持有者和计数。

3. **释放锁**：
    - 当请求释放锁时，如果它是当前持有者，减少计数。如果计数减到0，则完全释放锁，并删除持有者记录。

4. **续期锁**：
    - 续期操作仅在持有者是当前请求时有效，更新锁的过期时间。

### Lua 脚本实现

#### 1. 获取锁的 Lua 脚本

```lua
local lock_key = KEYS[1]
local holder_key = lock_key .. ':holder'
local count_key = lock_key .. ':count'
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

-- 检查当前锁持有者
local current_holder = redis.call('GET', holder_key)
local current_count = tonumber(redis.call('GET', count_key) or '0')

-- 如果锁未被持有，或者当前持有者是请求 ID，则获取锁
if not current_holder or current_holder == request_id then
    redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl)
    redis.call('SET', holder_key, request_id)
    redis.call('SET', count_key, current_count + 1)
    return 'OK'
end

return nil
```

#### 2. 释放锁的 Lua 脚本

```lua
local lock_key = KEYS[1]
local holder_key = lock_key .. ':holder'
local count_key = lock_key .. ':count'
local request_id = ARGV[1]

-- 检查当前锁持有者
local current_holder = redis.call('GET', holder_key)
local current_count = tonumber(redis.call('GET', count_key) or '0')

if current_holder == request_id then
    if current_count > 1 then
        redis.call('SET', count_key, current_count - 1)
        return 'OK'
    else
        redis.call('DEL', lock_key)
        redis.call('DEL', holder_key)
        redis.call('DEL', count_key)
        return 'OK'
    end
end

return nil
```

#### 3. 锁续期的 Lua 脚本

```lua
local lock_key = KEYS[1]
local holder_key = lock_key .. ':holder'
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

-- 检查当前锁持有者
local current_holder = redis.call('GET', holder_key)

if current_holder == request_id then
    redis.call('EXPIRE', lock_key, lock_ttl)
    return 'OK'
end

return nil
```

### 关键点

1. **持有者记录**：
    - `holder_key` 记录当前锁的持有者。如果当前持有者与请求 ID 匹配，则允许进行操作。

2. **计数管理**：
    - `count_key` 记录持有者持有锁的次数。这样，同一请求可以多次获得锁而不会死锁。每次获得锁时增加计数，每次释放锁时减少计数。

3. **锁过期时间**：
    - 在获取锁时设置过期时间。在续期操作中更新过期时间，以保持锁的有效性。

4. **原子性**：
    - 使用 Lua 脚本可以确保操作的原子性，避免在锁操作过程中发生竞态条件。

### 总结

可重入锁在分布式系统中可以通过 Redis 实现，利用 Redis 的键值对机制来管理锁的持有者和计数。通过 Lua 脚本的方式，能够保证锁操作的原子性，从而避免并发问题。持有者记录和计数管理是实现可重入锁的核心部分，而过期时间和续期机制则确保锁的有效性和可用性。这种方法适用于需要保证同一请求可以多次获得锁而不发生死锁的场景。
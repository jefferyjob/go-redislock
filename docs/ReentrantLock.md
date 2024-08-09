### 可重入锁的实现

#### 实现思路

可重入锁（Reentrant Lock）是一种锁，它允许同一个线程或进程多次获得同一把锁而不会发生死锁。这通常通过记录锁的持有者和持有次数来实现。

在分布式环境中，Redis 可以用来实现可重入锁。实现思路如下：

1. **锁的持有者和计数**：
    - 使用 Redis 键存储锁的持有者（通常是唯一标识符，如请求 ID 或令牌）。
    - 使用另一个键存储持有者的计数，表示当前持有锁的次数。

2. **获取锁**：
    - 如果当前没有持有者，则尝试获得锁，并记录持有者和计数。
    - 如果已经持有者是当前请求，增加计数。

3. **释放锁**：
    - 减少持有者的计数。
    - 如果计数减到 0，则释放锁并删除持有者记录。

4. **锁续期**：
    - 续期时仅在持有者是当前请求时有效，更新锁的过期时间。

### Lua 脚本实现

#### 1. 获取锁的 Lua 脚本 (`acquire_lock.lua`)

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

#### 2. 释放锁的 Lua 脚本 (`release_lock.lua`)

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

#### 3. 锁续期的 Lua 脚本 (`renew_lock.lua`)

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

1. **持有者管理**：
    - 使用 Redis 键 (`holder_key`) 来跟踪锁的持有者。每次请求获取锁时，都检查当前持有者是否是请求 ID。

2. **计数管理**：
    - 使用另一个 Redis 键 (`count_key`) 来管理持有者的计数。这可以确保同一持有者可以多次获得锁，而不会发生死锁。

3. **锁过期**：
    - 锁过期时间需要在获取锁时设置，并在续期时更新。如果锁的持有者尝试续期锁，只有当持有者是当前请求时，才会成功。

4. **锁释放**：
    - 释放锁时需要检查持有者和计数。如果计数大于 1，则减少计数；如果计数为 1，则完全释放锁并删除相关记录。

### 总结

可重入锁在分布式环境中可以通过 Redis 实现，利用 Redis 的键值对机制来管理锁的持有者和计数。关键在于确保只有持有锁的请求能够增加计数、释放锁和续期。使用 Lua 脚本可以保证操作的原子性，防止竞态条件和死锁的发生。通过这种方式，可以实现高效且可靠的分布式可重入锁。
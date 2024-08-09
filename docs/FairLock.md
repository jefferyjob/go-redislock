`fairLock.lua` 是用来实现公平锁的 Lua 脚本。它的主要目标是确保锁的请求是按照请求的顺序获得锁的，同时处理超时请求。以下是脚本的详细解释：

### Lua 脚本解释

```lua
local lock_key = KEYS[1]
local queue_key = lock_key .. ':queue'
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local request_timeout = tonumber(ARGV[3])

-- 清理超时的请求
local current_time = tonumber(redis.call('TIME')[1])
redis.call('ZREMRANGEBYSCORE', queue_key, 0, current_time - request_timeout)

-- 将请求 ID 添加到队列中，并设置过期时间
redis.call('ZADD', queue_key, current_time, request_id)
redis.call('EXPIRE', queue_key, request_timeout)

-- 检查当前请求是否为队列的第一个
local first_request_id = redis.call('ZRANGE', queue_key, 0, 0, 'WITHSCORES')[1]

if request_id == first_request_id then
    -- 如果当前请求是第一个，尝试获得锁
    if redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl) then
        return 'OK'
    end
end

return nil
```

### 关键步骤解释

1. **清理超时的请求**

   ```lua
   local current_time = tonumber(redis.call('TIME')[1])
   redis.call('ZREMRANGEBYSCORE', queue_key, 0, current_time - request_timeout)
   ```

    - `redis.call('TIME')[1]` 获取当前的时间（秒数）。
    - `redis.call('ZREMRANGEBYSCORE', queue_key, 0, current_time - request_timeout)` 从排序集合中删除所有超时的请求。`ZREMRANGEBYSCORE` 是 Redis 的一个命令，用于从有序集合中删除指定分数范围的成员。在这里，`0` 到 `current_time - request_timeout` 的范围表示请求时间早于当前时间减去请求超时的阈值（即超时的请求）。

2. **添加当前请求到队列中**

   ```lua
   redis.call('ZADD', queue_key, current_time, request_id)
   redis.call('EXPIRE', queue_key, request_timeout)
   ```

    - `redis.call('ZADD', queue_key, current_time, request_id)` 将当前请求 ID 和时间（`current_time`）添加到有序集合中。这个操作将当前请求放入队列，并按时间排序。
    - `redis.call('EXPIRE', queue_key, request_timeout)` 设置队列的过期时间，这样它不会无限期存在。

3. **检查当前请求是否为队列中的第一个**

   ```lua
   local first_request_id = redis.call('ZRANGE', queue_key, 0, 0, 'WITHSCORES')[1]
   
   if request_id == first_request_id then
       if redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl) then
           return 'OK'
       end
   end
   ```

    - `redis.call('ZRANGE', queue_key, 0, 0, 'WITHSCORES')[1]` 从队列中获取第一个请求 ID（分数最低的请求）。`ZRANGE` 命令用于从有序集合中获取指定范围的成员，这里我们只获取第一个成员。
    - 如果当前请求 ID 是队列中的第一个请求 ID，尝试设置锁（`redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl)`）。如果设置成功，则当前请求获得锁。

### 总结

- `ZREMRANGEBYSCORE` 命令的作用是清除超时的请求，确保队列中只保留有效的请求。
- 该脚本保证了公平性，即请求按照其到达的顺序获得锁，并且如果请求超时，它们会被从队列中清除，以免影响后续请求。

这种实现方式确保了锁请求的公平性，并且处理了超时的情况，从而避免了队列无限增长的问题。
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

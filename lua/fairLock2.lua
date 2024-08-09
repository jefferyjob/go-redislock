local lock_key = KEYS[1]
local lock_queue_key = lock_key .. ':queue'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local request_id = ARGV[3]
local max_wait_time = tonumber(ARGV[4]) -- 最大等待时间

-- 将请求ID加入队列末尾
redis.call('LPUSH', lock_queue_key, request_id)

-- 检查队列头部是否是当前请求的ID
if redis.call('LINDEX', lock_queue_key, -1) == request_id then
    -- 检查锁是否未被持有
    if redis.call('SET', lock_key, lock_value, 'NX', 'EX', lock_ttl) then
        redis.call('RPOP', lock_queue_key)
        return "OK"
    end
end

-- 检查最大等待时间，移除超时的请求
local request_time_str = redis.call('LINDEX', lock_queue_key, -1)
if request_time_str then
    local request_time = tonumber(request_time_str:match("^(%d+)"))
    if request_time and (redis.call('TIME')[1] - request_time) > max_wait_time then
        redis.call('RPOP', lock_queue_key)
    end
end

return nil


--推荐使用
local lock_key = KEYS[1]
local queue_key = lock_key .. ':queue'
local request_id = ARGV[1]

-- 如果锁的持有者 ID 与当前请求 ID 匹配
if redis.call('GET', lock_key) == request_id then
    -- 删除锁
    redis.call('DEL', lock_key)

    -- 从队列中移除当前 ID
    redis.call('LREM', queue_key, 0, request_id)

    -- 尝试将队列中的下一个请求设置为锁的持有者
    local next_request_id = redis.call('LRANGE', queue_key, 0, 0)[1]
    if next_request_id then
        redis.call('SET', lock_key, next_request_id, 'NX', 'EX', redis.call('TTL', lock_key))
    end

    return 'OK'
end

return nil

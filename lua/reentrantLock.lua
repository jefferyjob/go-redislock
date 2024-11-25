local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

-- 可重入锁计数器
if reentrant_count > 0 then
    redis.call('INCR', reentrant_key)
    redis.call('PEXPIRE', lock_key, lock_ttl)
    redis.call('PEXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

-- 创建锁
if redis.call('SET', lock_key, lock_value, 'NX', 'PX', lock_ttl) then
    redis.call('SET', reentrant_key, 1)
    redis.call('PEXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

return nil

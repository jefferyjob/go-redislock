local lock_key = KEYS[1]
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

if reentrant_count > 0 then
    redis.call('INCR', reentrant_key)
    redis.call('EXPIRE', lock_key, lock_ttl)
    redis.call('EXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

if redis.call('SET', lock_key, lock_value, 'NX', 'EX', lock_ttl) then
    redis.call('SET', reentrant_key, 1)
    redis.call('EXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

return nil
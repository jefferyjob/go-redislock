local lock_key = KEYS[1]
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

if reentrant_count > 0 or redis.call('GET', lock_key) == lock_value then
    redis.call('EXPIRE', lock_key, lock_ttl)
    redis.call('EXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

return nil
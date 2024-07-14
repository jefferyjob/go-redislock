local lock_key = KEYS[1]
local lock_value = ARGV[1]
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

if reentrant_count > 1 then
    redis.call('DECR', reentrant_key)
    return "OK"
elseif reentrant_count == 1 then
    redis.call('DEL', reentrant_key)
    redis.call('DEL', lock_key)
    return "OK"
end

if redis.call('GET', lock_key) == lock_value then
    redis.call('DEL', lock_key)
    return "OK"
else
    return nil
end
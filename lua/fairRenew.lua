local lock_key = KEYS[1]
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

if redis.call('GET', lock_key) == request_id then
    redis.call('EXPIRE', lock_key, lock_ttl)
    return 'OK'
end

return nil

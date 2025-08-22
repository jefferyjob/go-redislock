local write_key = KEYS[1]
local token = ARGV[1]
local ttl = tonumber(ARGV[2])

local current_write = redis.call('GET', write_key)
if (current_write ~= false and current_write ~= token) then
    return 0
end

local read_counter_key = write_key .. ':read:' .. token
redis.call('INCR', read_counter_key)
redis.call('PEXPIRE', read_counter_key, ttl)

-- 记录全局读锁数量
redis.call('INCR', write_key .. ':read_count')
redis.call('PEXPIRE', write_key .. ':read_count', ttl)

return 1

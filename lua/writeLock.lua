local write_key = KEYS[1]
local token = ARGV[1]
local ttl = tonumber(ARGV[2])

local read_count = tonumber(redis.call('GET', write_key .. ':read_count') or '0')
local current_write = redis.call('GET', write_key)

if (read_count == 0 and (current_write == false or current_write == token)) then
    redis.call('SET', write_key, token, 'PX', ttl)
    local counter_key = write_key .. ':write:' .. token
    redis.call('INCR', counter_key)
    redis.call('PEXPIRE', counter_key, ttl)
    return 1
end

return 0

local write_key = KEYS[1]
local token = ARGV[1]

if redis.call('GET', write_key) ~= token then
    return 0
end

local counter_key = write_key .. ':write:' .. token
local count = tonumber(redis.call('GET', counter_key) or '0')

if count <= 1 then
    redis.call('DEL', write_key)
    redis.call('DEL', counter_key)
else
    redis.call('DECR', counter_key)
end

return 1

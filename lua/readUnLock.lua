local write_key = KEYS[1]
local token = ARGV[1]

local read_counter_key = write_key .. ':read:' .. token
local count = tonumber(redis.call('GET', read_counter_key) or '0')

if count <= 1 then
    redis.call('DEL', read_counter_key)
else
    redis.call('DECR', read_counter_key)
end

-- 减少全局读锁数量
local total_reads = tonumber(redis.call('DECR', write_key .. ':read_count'))
if total_reads <= 0 then
    redis.call('DEL', write_key .. ':read_count')
end

return 1

-- 锁 key 和 value
local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

-- 尝试创建锁
if redis.call('SET', lock_key, lock_value, 'NX', 'PX', lock_ttl) then
    return 1
end

-- 获取失败
return 0

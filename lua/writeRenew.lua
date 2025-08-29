local local_key = KEYS[1]
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2]) or 0

-- 验证写锁持有者
local writer = redis.call('HGET', local_key, 'writer')
if writer ~= lock_value then
    -- 非写锁持有者，续期失败
    return 0
end

-- 刷新 TTL
redis.call('PEXPIRE', local_key, lock_ttl)
return 1
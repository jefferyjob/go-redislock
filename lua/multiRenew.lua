-- 锁 key 和持有者标识
local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])

-- 只有持有锁的客户端才能续期
if redis.call('GET', lock_key) == lock_value then
    redis.call('PEXPIRE', lock_key, lock_ttl)
    return 1
end

-- 续期失败（锁不存在或不是持有者）
return 0

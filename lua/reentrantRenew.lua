local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

-- 锁续期
-- 重入锁的场景（reentrant_count > 0）
-- 普通锁的场景（redis.call('GET', lock_key) == lock_value）
if reentrant_count > 0 or redis.call('GET', lock_key) == lock_value then
    redis.call('PEXPIRE', lock_key, lock_ttl)
    redis.call('PEXPIRE', reentrant_key, lock_ttl)
    return "OK"
end

return nil

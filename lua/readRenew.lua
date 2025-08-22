local local_key = KEYS[1]
local lock_value = ARGV[1] -- 当前请求续期的持有者标识（owner）
local lock_ttl = tonumber(ARGV[2]) or 0

-- 获取自身读锁计数，判断是否持有读锁
local self_cnt = tonumber(redis.call('HGET', local_key, 'r:' .. lock_value) or '0')

-- 如果当前线程没有持有读锁，则续期失败
if self_cnt <= 0 then
    return 0
end

-- 刷新锁的 TTL，延长锁有效期，避免锁过期被其他线程抢占
redis.call('PEXPIRE', local_key, lock_ttl)
return 1
-- 锁 key 和持有者标识
local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]

-- 只有持有锁的客户端才能释放
if redis.call('GET', lock_key) == lock_value then
    redis.call('DEL', lock_key)
    return 1
end

-- 解锁失败（锁不存在或不是持有者）
return 0

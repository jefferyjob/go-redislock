local local_key = KEYS[1]
local lock_value = ARGV[1] -- 当前请求解锁的持有者标识（owner）

-- 获取当前写锁持有者
local writer = redis.call('HGET', local_key, 'writer')
if writer ~= lock_value then
    -- 如果当前线程不是写锁持有者，则解锁失败
    return 0
end

-- 减少写锁计数（支持可重入锁）
local wcount = tonumber(redis.call('HINCRBY', local_key, 'wcount', -1))
if wcount > 0 then
    -- 写锁仍然持有（可重入计数 > 0），无需释放锁，直接返回
    return 1
end

-- 写锁计数归零，释放写锁
redis.call('HDEL', local_key, 'writer')
redis.call('HDEL', local_key, 'wcount')

-- 检查是否存在读锁
local rcount = tonumber(redis.call('HGET', local_key, 'rcount') or '0')
if rcount > 0 then
    -- 仍有读锁，切回读模式
    redis.call('HSET', local_key, 'mode', 'read')
else
    -- 无锁持有者，删除键
    redis.call('DEL', local_key)
end

return 1
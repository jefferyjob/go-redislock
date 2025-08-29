local local_key = KEYS[1]
local lock_value = ARGV[1] -- 当前请求锁的持有者标识（owner）
local lock_ttl = tonumber(ARGV[2]) or 0

-- 获取当前锁模式
local mode = redis.call('HGET', local_key, 'mode')

if not mode then
    -- 如果锁不存在（空闲状态），直接加读锁
    -- 初始化锁信息：
    -- mode: 'read' 表示读锁模式
    -- rcount: 当前读锁总数
    -- r:{owner}: 当前持有者的读锁计数（可重入）
    redis.call('HSET', local_key,
        'mode', 'read',
        'rcount', 1,
        'r:' .. lock_value, 1
    )
    redis.call('PEXPIRE', local_key, lock_ttl)
    return 1
end

-- 如果当前锁模式是读锁
if mode == 'read' then
    -- 读读并发：累加本 lock_value 的读计数与总读者数

    -- 多个读者可并发获取锁
    -- 增加当前持有者的读锁计数
    redis.call('HINCRBY', local_key, 'r:' .. lock_value, 1)
    -- 增加读的计数
    redis.call('HINCRBY', local_key, 'rcount', 1)
    -- 刷新 TTL
    redis.call('PEXPIRE', local_key, lock_ttl)
    return 1
end


-- 如果当前锁模式是写锁
local writer = redis.call('HGET', local_key, 'writer')
if writer == lock_value then
    -- 自己持有写锁：允许同时持有读锁（读写可重入）

    -- 如果当前持有者就是写锁的持有者，则允许读写可重入
    -- 增加当前持有者的读锁计数
    redis.call('HINCRBY', local_key, 'r:' .. lock_value, 1)
    -- 增加总读者数
    redis.call('HINCRBY', local_key, 'rcount', 1)
    -- 刷新 TTL
    redis.call('PEXPIRE', local_key, lock_ttl)
    return 1
end


-- 如果锁是写锁且持有者不是自己，则无法获取读锁
-- 他人持有写锁：失败
return 0
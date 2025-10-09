local local_key = KEYS[1]
local lock_value = ARGV[1] -- 当前请求锁的持有者标识（owner）
local lock_ttl = tonumber(ARGV[2]) or 0

-- 获取当前锁模式
local mode = redis.call('HGET', local_key, 'mode')

if not mode then
    -- 如果锁不存在（空闲状态），直接加写锁
    -- 初始化锁信息：
    -- mode: 'write' 表示写锁模式
    -- writer: 当前写锁持有者
    -- wcount: 写锁可重入计数
    redis.call('HSET', local_key,
        'mode', 'write',
        'writer', lock_value,
        'wcount', 1)
    -- 设置锁过期时间，避免死锁
    redis.call('PEXPIRE', local_key, lock_ttl)
    return 1
end

-- 如果当前锁模式是写锁
if mode == 'write' then
    -- 获取写锁持有者
    local writer = redis.call('HGET', local_key, 'writer')
    if writer == lock_value then
        -- 可重入写锁
        -- 当前持有者再次请求写锁，可重入
        redis.call('HINCRBY', local_key, 'wcount', 1)
        -- 刷新 TTL
        redis.call('PEXPIRE', local_key, lock_ttl)
        return 1
    else
        -- 他人持有写锁，获取失败
        return 0
    end
end

-- 如果当前锁模式是读锁
if mode == 'read' then
    -- 总读者数
    local total = tonumber(redis.call('HGET', local_key, 'rcount') or '0')
    -- 自己的读锁计数
    local self_cnt = tonumber(redis.call('HGET', local_key, 'r:' .. lock_value) or '0')
    if total == self_cnt then
        -- 仅自己持有读锁，可以升级为写锁
        redis.call('HSET', local_key,
                'mode', 'write',
                'writer', lock_value,
                'wcount', 1)
        redis.call('PEXPIRE', local_key, lock_ttl)
        return 1
    end
end


-- 其他情况无法获取写锁（存在其他读者或写锁被他人占用）
return 0
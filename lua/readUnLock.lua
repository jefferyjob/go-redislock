--[[
    Read Lock Unlock Script (读锁解锁脚本)

    输入参数：
    KEYS[1]     - 锁的业务 key（Hash）
    ARGV[1]     - 当前请求解锁的持有者标识（owner）

    返回值：
    - 1：本持有者的读锁已完全释放
    - 2 ：部分解锁（可重入计数减 1，本持有者仍持有读锁）
    - 0 ：解锁失败（本持有者未持有读锁）
--]]

local local_key = KEYS[1]
local lock_value = ARGV[1] -- 当前请求解锁的持有者标识（owner）

-- 获取当前持有者的读锁计数
local self_cnt = tonumber(redis.call('HGET', local_key, 'r:' .. lock_value) or '0')

-- 如果自身没有持有读锁，则解锁失败
if self_cnt <= 0 then
    return 0
end

-- 减少自身读锁计数
self_cnt = redis.call('HINCRBY', local_key, 'r:' .. lock_value, -1)
-- 自身读锁减 1
redis.call('HINCRBY', local_key, 'rcount', -1)

-- 可重入：自身仍持有读锁，返回 2（保留 TTL 续期等持锁状态）
if self_cnt > 0 then
    return 2
end

-- 如果自身读锁计数归零，删除自身读锁字段
redis.call('HDEL', local_key, 'r:' .. lock_value)

-- 获取总读者数
local total = tonumber(redis.call('HGET', local_key, 'rcount') or '0')
if total <= 0 then
    -- 当没有其他读者时，需要根据模式判断是否清理键
    local mode = redis.call('HGET', local_key, 'mode') -- 当前锁模式
    if mode == 'read' then
        -- 如果当前模式是读锁，且总读者数为 0，则删除整个锁键
        redis.call('DEL', local_key)
    else
        -- 如果模式是写锁，说明还有写锁存在，只删除读锁计数字段
        redis.call('HDEL', local_key, 'rcount')
    end
end

return 1

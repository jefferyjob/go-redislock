--[[
    Fair Queue Distributed Unlock Script (基于排队公平锁的解锁脚本)

    功能描述：
    用于释放通过公平排队机制获取的分布式锁。确保只有当前请求者才有权限释放锁，
    并从 Redis ZSET 队列中移除自己的请求 ID，避免长时间占用排队资源。

    使用场景：
    搭配基于 ZSET 实现的公平分布式锁脚本使用，客户端释放锁时调用。

    输入参数：
    KEYS[1]      - 业务锁 key（如 "my-lock"）
    ARGV[1]      - 请求 ID（客户端持锁标识，建议与加锁时传入一致）

    Redis 数据结构说明：
    1. 主锁键（{KEYS[1]}）：存储当前持锁请求 ID；
    2. 排队键（{KEYS[1]}:queue）：ZSET，记录所有等待请求，score 为时间戳。

    执行逻辑：
    1. 若当前请求 ID 与锁键中的值一致（是锁的持有者），则删除锁键并记为解锁成功；
    2. 无论是否持有锁，统一从 ZSET 排队队列中移除该请求 ID；
    3. 返回是否真正释放了锁。

    返回值：
    - 1：解锁成功（当前请求是持有者，锁键已删除）
    - 0 ：解锁失败（锁不存在或当前请求不是持有者）；排队记录仍会被移除

    注意事项：
    - 使用 `GET lock_key == request_id` 判断是否是锁的持有者；
    - 解锁时必须确保 request_id 和加锁时保持一致；
    - 队列清理（ZREM）是幂等的，多次调用不会产生副作用；
    - 与基于 ZSET 的加锁脚本配套使用效果最佳。
--]]


local lock_key = '{' .. KEYS[1] .. '}'
local queue_key = lock_key .. ':queue'
local request_id = ARGV[1]

local unlocked = 0
-- 删除锁键（只删除自己持有的锁）
if redis.call('GET', lock_key) == request_id then
    redis.call('DEL', lock_key)
    unlocked = 1
end

-- 从队列中删除请求ID
redis.call('ZREM', queue_key, request_id)

return unlocked

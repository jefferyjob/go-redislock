--[[
    Reentrant Distributed Lock TTL Renewal Script (可重入分布式锁续期脚本)

    功能描述：
    该 Lua 脚本用于对 Redis 中的分布式锁进行续期（刷新 TTL），
    支持可重入场景和普通锁场景。仅当当前客户端仍然持有锁时，才会执行续期操作。

    使用场景：
    在持锁期间业务执行时间较长，需要在锁到期前定期刷新锁的 TTL，
    防止锁因过期被其他客户端抢占导致数据竞争或并发问题。

    输入参数：
    KEYS[1]     - 锁的业务 key（如 "my-lock"）
    ARGV[1]     - 当前客户端标识（如 UUID，作为 lock_value）
    ARGV[2]     - 续期的 TTL（单位：毫秒）

    Redis 数据结构：
    1. 主锁 key:
        格式：{KEYS[1]}
        值：ARGV[1]（客户端标识）
    2. 可重入计数器 key:
        格式：{KEYS[1]}:count:{ARGV[1]}
        值：整数，表示当前客户端的重入次数

    执行逻辑：
    1. 读取当前客户端对应的可重入计数器（reentrant_key）；
    2. 满足以下任意一个条件即认为客户端持有锁，可以续期：
        - 可重入计数器存在（reentrant_count > 0）
        - 主锁值等于当前客户端的标识（redis.call('GET', lock_key) == lock_value）
    3. 若满足续期条件：
        - 刷新主锁和重入计数器的过期时间（PEXPIRE）；
        - 返回 1 表示续期成功；
    4. 否则（锁不存在或不是当前客户端持有），返回 nil。

    返回值：
    - 1：续期成功（当前客户端仍持有锁）
    - 0 ：续期失败（锁不存在或非本客户端持有）

    注意事项：
    - 客户端应定时调用该脚本以实现“自动续租”功能；
    - 续期操作需要和加锁、解锁脚本配套使用，并保持客户端 lock_value 一致；
    - 锁 key 和重入计数器 key 使用 Redis hash tag `{}` 包裹，确保在 Redis Cluster 下分布在同一 slot。

--]]


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
    return 1
end

return 0

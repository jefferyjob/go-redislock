--[[
    Reentrant Distributed Lock Script (可重入分布式锁脚本)

    功能描述：
    该 Lua 脚本用于在 Redis 中实现一个支持可重入的分布式锁机制。
    支持同一个客户端（以 lock_value 标识）在持有锁的情况下再次获取锁而不会阻塞或失败。

    使用场景：
    在分布式系统中，需要确保某段关键逻辑同一时间只能由一个客户端执行；
    且客户端可能递归或多次尝试加锁，需要支持可重入。

    输入参数：
    KEYS[1]     - 业务锁的主键名（如 "my-lock"）
    ARGV[1]     - 当前客户端标识（如 UUID，作为锁值 lock_value）
    ARGV[2]     - 锁的过期时间（单位：毫秒，lock_ttl）

    Redis 数据结构：
    1. 主锁 key:
        格式：{KEYS[1]}
        值：ARGV[1]（客户端 ID）
        设置：SET NX PX lock_ttl
    2. 可重入计数器 key:
        格式：{KEYS[1]}:count:{ARGV[1]}
        值：整数，表示客户端当前持有锁的重入次数

    执行逻辑：
    1. 首先尝试读取客户端自己的可重入计数器；
        - 如果大于 0，说明该客户端已经持有锁：
            - 将重入计数加 1；
            - 刷新主锁和计数器的过期时间；
            - 返回 1，表示加锁成功。
    2. 如果计数器不存在或为 0，说明该客户端未持有锁：
        - 尝试使用 SET NX PX 加锁；
        - 如果成功，设置可重入计数器为 1，并设置过期时间；
        - 返回 1，表示加锁成功。
    3. 如果 SET NX 加锁失败，表示已有其他客户端持有锁：
        - 返回 0，表示加锁失败。

    返回值：
    - 1：加锁成功（首次或可重入）
    - 0 ：加锁失败（被其他客户端持有）

    注意事项：
    - 锁名（KEYS[1]）应使用 Redis 的 hash tag `{}` 包裹，确保主锁和重入计数器落在同一 slot（用于 Redis Cluster）。
    - 客户端释放锁时，需正确管理可重入次数递减并在计数为 0 时删除主锁和计数器。
--]]


local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

-- 可重入锁计数器
if reentrant_count > 0 then
    redis.call('INCR', reentrant_key)
    redis.call('PEXPIRE', lock_key, lock_ttl)
    redis.call('PEXPIRE', reentrant_key, lock_ttl)
    return 1
end

-- 创建锁
if redis.call('SET', lock_key, lock_value, 'NX', 'PX', lock_ttl) then
    redis.call('SET', reentrant_key, 1)
    redis.call('PEXPIRE', reentrant_key, lock_ttl)
    return 1
end

return 0

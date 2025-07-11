--[[
    Reentrant Distributed Unlock Script (可重入分布式锁释放脚本)

    功能描述：
    该 Lua 脚本用于在 Redis 中释放一个支持可重入的分布式锁。
    如果客户端持有锁多次（重入），则需要多次调用此脚本释放；只有最后一次释放时，才会真正删除主锁 key。

    使用场景：
    与加锁脚本配合，用于支持分布式场景下的可重入加锁/解锁逻辑。
    确保只有加锁的客户端才能解锁，并正确处理重入计数。

    输入参数：
    KEYS[1]     - 锁的业务 key（如 "my-lock"）
    ARGV[1]     - 当前客户端标识（如 UUID，作为 lock_value）

    Redis 数据结构：
    1. 主锁 key:
        格式：{KEYS[1]}
        值：ARGV[1]（客户端 ID）
    2. 可重入计数器 key:
        格式：{KEYS[1]}:count:{ARGV[1]}
        值：整数，表示该客户端的持锁次数

    执行逻辑：
    1. 构造锁名 lock_key 和可重入计数器名 reentrant_key；
    2. 如果 reentrant_count > 1：
        - 表示客户端还持有多次锁，仅减 1 并返回 1；
    3. 如果 reentrant_count == 1：
        - 删除计数器；
        - 如果主锁的值等于客户端标识，则删除主锁；
        - 返回 1；
    4. 如果计数器不存在或为 0：
        - 尝试作为普通非重入锁解锁；
        - 如果主锁的值等于客户端标识，则删除主锁，返回 1；
    5. 以上均不满足（如不是持有者），返回 0 表示解锁失败。

    返回值：
    - 1：解锁成功（无论是否重入）
    - 0 ：解锁失败（锁不存在或当前客户端不是持有者）

    注意事项：
    - 加锁和解锁脚本必须搭配使用，并保持客户端 lock_value 一致；
    - lock_key 使用 Redis hash tag `{}` 包裹，确保与重入计数器位于同一 slot（用于 Redis Cluster）；
    - 客户端需要确保在业务完成后调用解锁脚本，否则会造成死锁。

--]]


local lock_key = '{' .. KEYS[1] .. '}'
local lock_value = ARGV[1]
local reentrant_key = lock_key .. ':count:' .. lock_value
local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')

--可重入锁解锁
if reentrant_count > 1 then
    redis.call('DECR', reentrant_key)
    return 1
elseif reentrant_count == 1 then
    redis.call('DEL', reentrant_key)

    -- 如果锁的值相等，删除锁
    if redis.call('GET', lock_key) == lock_value then
        redis.call('DEL', lock_key)
        return 1
    end
end

--非可重入锁解锁
if redis.call('GET', lock_key) == lock_value then
    redis.call('DEL', lock_key)
    return 1
end

return 0

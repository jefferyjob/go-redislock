--[[
    Fair Queue Distributed Lock Using ZSET (基于有序集合的公平分布式锁脚本)

    功能描述：
    本 Lua 脚本使用 Redis ZSET 实现一个带排队机制的公平分布式锁。
    每个客户端以请求 ID 加入队列，并按时间顺序排队，只有队首请求才能尝试加锁。
    同时支持设置最大请求等待时间，自动清理过期请求，避免死锁或长时间占用队列。

    使用场景：
    - 多客户端需要公平排队获取同一个资源的场景；
    - 需要显式控制锁等待时间，清除僵尸请求；
    - 比传统 SETNX 更具公平性和可控性。

    输入参数：
    KEYS[1]      - 锁的 key（如 "resource-lock"）
    ARGV[1]      - 请求 ID（一般为客户端 ID + 唯一请求标识，如 UUID）
    ARGV[2]      - 锁的过期时间（秒，lock_ttl）
    ARGV[3]      - 请求最大等待时间（秒，request_timeout）

    Redis 数据结构说明：
    1. 锁 key:     Redis String，存储当前持有锁的请求 ID
    2. 排队 key:   Redis Sorted Set（ZSET），score 为请求时间戳，value 为请求 ID

    执行流程：
    1. 获取当前时间戳 current_time；
    2. 清理 queue_key 中所有超过 request_timeout 的请求（ZREMRANGEBYSCORE）；
    3. 将当前请求 ID 按当前时间戳添加到 ZSET 队列中（ZADD）；
    4. 设置 queue_key 过期时间为 request_timeout（用于自动过期清理）；
    5. 检查当前请求是否是队首（ZRANGE 0 0）：
        - 是，则尝试使用 SET NX EX 获取锁；
        - 如果成功，加锁成功，返回 1；
        - 否则或不是队首，则返回 0。

    返回值：
    - 1：加锁成功（当前请求是队首且成功获取锁）
    - 0 ：加锁失败（未轮到或抢锁失败）

    建议使用说明（客户端逻辑）：
    - 客户端加锁失败应设置间隔轮询重试；
    - 请求 ID 应具备唯一性；
    - 可扩展为锁续期、解锁和队列清理等完整锁管理模块。

    注意事项：
    - 当前时间使用 `redis.call('TIME')[1]`，单位为秒；
    - 脚本设计为幂等，重复调用不会产生副作用；
    - 若客户端意外宕机未解锁，锁将在 TTL 后自动释放，但队列中残留项会自动过期清除；
    - 可根据需要改为毫秒时间戳并配合 PEXPIRE 精细控制。

--]]


local lock_key = '{' .. KEYS[1] .. '}'
local queue_key = lock_key .. ':queue'
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local request_timeout = tonumber(ARGV[3])

-- 清理超时的请求
local current_time = tonumber(redis.call('TIME')[1])
redis.call('ZREMRANGEBYSCORE', queue_key, 0, current_time - request_timeout)

-- 加锁（排队）
-- 将请求 ID 添加到队列中，并设置过期时间
redis.call('ZADD', queue_key, current_time, request_id)
redis.call('EXPIRE', queue_key, request_timeout)

-- 检查当前请求是否为队列的第一个
local first_request_id = redis.call('ZRANGE', queue_key, 0, 0, 'WITHSCORES')[1]

-- 获得锁（只有队首才能成功获得锁）
if request_id == first_request_id then
    -- 如果当前请求是第一个，尝试获得锁
    if redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl) then
        return 1
    end
end

return 0
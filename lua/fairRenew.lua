--[[
    Fair Lock Renew TTL   (公平锁锁续期脚本)

    适用场景：
      - 客户端在业务处理中需要延长持锁时间，防止锁因 TTL 到期提前释放。

    KEYS[1]      - 锁的 key（与加锁脚本保持一致）
    ARGV[1]      - 请求 ID（必须与加锁时写入的完全一致）
    ARGV[2]      - 续期的锁 TTL（毫秒，lock_ttl）

    返回：
      1  续期成功（确实持有该锁并已刷新 TTL）
      0  续期失败（锁不存在，或锁已被其他请求持有）
--]]


local lock_key      = '{' .. KEYS[1] .. '}'
local request_id    = ARGV[1]
local lock_ttl  = tonumber(ARGV[2])

-- 只允许当前持锁者续期
if redis.call('GET', lock_key) == request_id then
    redis.call('PEXPIRE', lock_key, lock_ttl)
    return 1 -- 续期成功
end

return 0  -- 续期失败：要么锁不存在，要么不是你的
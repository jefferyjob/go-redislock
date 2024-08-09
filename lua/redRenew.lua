local lock_keys = KEYS
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local renewed = {}

-- 尝试在所有 Redis 实例上续期
for i, lock_key in ipairs(lock_keys) do
    local current_value = redis.call('GET', lock_key)
    if current_value == request_id then
        redis.call('EXPIRE', lock_key, lock_ttl)
        table.insert(renewed, lock_key)
    else
        -- 如果续期失败，则释放所有已续期的锁
        for _, key in ipairs(renewed) do
            redis.call('DEL', key)
        end
        return nil
    end
end

return 'OK'

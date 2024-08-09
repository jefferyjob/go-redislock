local lock_keys = KEYS
local request_id = ARGV[1]
local lock_ttl = tonumber(ARGV[2])
local acquired = {}

-- 尝试获取所有锁
for i, lock_key in ipairs(lock_keys) do
    local result = redis.call('SET', lock_key, request_id, 'NX', 'EX', lock_ttl)
    if result == 'OK' then
        table.insert(acquired, lock_key)
    else
        -- 如果获取锁失败，则释放所有已成功获取的锁
        for _, key in ipairs(acquired) do
            redis.call('DEL', key)
        end
        return nil
    end
end

return 'OK'

local lock_keys = KEYS
local request_id = ARGV[1]

-- 尝试释放所有锁
for _, lock_key in ipairs(lock_keys) do
    if redis.call('GET', lock_key) == request_id then
        redis.call('DEL', lock_key)
    end
end

return 'OK'

package main

const (
	// 加锁
	lockScript = `
		local lock_key = KEYS[1]
		local lock_value = ARGV[1]
		local lock_ttl = tonumber(ARGV[2])
		local reentrant_key = lock_key .. ':count:' .. lock_value
		local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')
		
		if reentrant_count > 0 then
			redis.call('INCR', reentrant_key)
			redis.call('EXPIRE', lock_key, lock_ttl)
			redis.call('EXPIRE', reentrant_key, lock_ttl) 
			return "OK"
		end
		
		if redis.call('SET', lock_key, lock_value, 'NX', 'EX', lock_ttl) then
			redis.call('SET', reentrant_key, 1)
			redis.call('EXPIRE', reentrant_key, lock_ttl) 
			return "OK"
		end
		
		return nil
	`

	// 解锁
	unLockScript = `
		local lock_key = KEYS[1]
		local lock_value = ARGV[1]
		local reentrant_key = lock_key .. ':count:' .. lock_value
		local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')
		
		if reentrant_count > 1 then
			redis.call('DECR', reentrant_key)
			return "OK"
		elseif reentrant_count == 1 then
			redis.call('DEL', reentrant_key)
			redis.call('DEL', lock_key)
			return "OK"
		end
		
		if redis.call('GET', lock_key) == lock_value then
			redis.call('DEL', lock_key)
			return "OK"
		else
			return nil
		end
	`

	// 续期
	renewScript = `
		local lock_key = KEYS[1]
		local lock_value = ARGV[1]
		local lock_ttl = tonumber(ARGV[2])
		local reentrant_key = lock_key .. ':count:' .. lock_value
		local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')
		
		if reentrant_count > 0 or redis.call('GET', lock_key) == lock_value then
			redis.call('EXPIRE', lock_key, lock_ttl)
			redis.call('EXPIRE', reentrant_key, lock_ttl)
			return "OK"
		end
		
		return nil
	`
)

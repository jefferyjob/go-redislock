package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"time"
)

var (
	//go:embed lua/writeLock.lua
	writeLockScript string
	//go:embed lua/writeUnLock.lua
	writeUnLockScript string
	//go:embed lua/writeRenew.lua
	writeRenewScript string
)

func (l *RedisLock) WLock(ctx context.Context) error {
	res, err := l.redis.Eval(ctx, writeLockScript, []string{l.key}, l.token, l.lockTimeout.Milliseconds()).Int64()
	if err != nil {
		return errors.Join(err, ErrException)
	}
	if res != 1 {
		return ErrLockFailed
	}
	return nil
}

func (l *RedisLock) WUnLock(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) SpinWLock(ctx context.Context, timeout time.Duration) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) WRenew(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

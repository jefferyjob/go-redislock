package go_redislock

import (
	"context"
	_ "embed"
	"errors"
	"time"
)

var (
	//go:embed lua/readLock.lua
	readLockScript string
	//go:embed lua/readUnLock.lua
	readUnLockScript string
	//go:embed lua/readRenew.lua
	readRenewScript string
)

func (l *RedisLock) RLock(ctx context.Context) error {
	res, err := l.redis.Eval(ctx, readLockScript, []string{l.key}, l.token, l.lockTimeout.Milliseconds()).Int64()
	if err != nil {
		return errors.Join(err, ErrException)
	}
	if res != 1 {
		return ErrLockFailed
	}
	return nil
}

func (l *RedisLock) RUnLock(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) SpinRLock(ctx context.Context, timeout time.Duration) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) RRenew(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

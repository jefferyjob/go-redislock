package go_redislock

import (
	"context"
	_ "embed"
	"time"
)

var (
	//go:embed lua/multiLock.lua
	multiLockScript string
	//go:embed lua/multiUnLock.lua
	multiUnLockScript string
	//go:embed lua/multiRenew.lua
	multiRenewScript string
)

func (l *RedisLock) MultiLock(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) MultiUnLock(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) SpinMultiLock(ctx context.Context, locks []RedisLockInter, timeout time.Duration) error {
	// TODO implement me
	panic("implement me")
}

func (l *RedisLock) MultiRenew(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	panic("implement me")
}

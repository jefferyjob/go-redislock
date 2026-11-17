package go_redislock

import (
	"context"
	_ "embed"
	"time"
)

var (
	// MultiLock 尚未实现，Lua 脚本先行保留备用。
	//go:embed lua/multiLock.lua
	multiLockScript string //nolint:unused // MultiLock 未实现，脚本保留备用
	//go:embed lua/multiUnLock.lua
	multiUnLockScript string //nolint:unused // MultiLock 未实现，脚本保留备用
	//go:embed lua/multiRenew.lua
	multiRenewScript string //nolint:unused // MultiLock 未实现，脚本保留备用
)

func (l *RedisLock) MultiLock(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	return ErrNotImplemented
}

func (l *RedisLock) MultiUnLock(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	return ErrNotImplemented
}

func (l *RedisLock) SpinMultiLock(ctx context.Context, locks []RedisLockInter, timeout time.Duration) error {
	// TODO implement me
	return ErrNotImplemented
}

func (l *RedisLock) MultiRenew(ctx context.Context, locks []RedisLockInter) error {
	// TODO implement me
	return ErrNotImplemented
}

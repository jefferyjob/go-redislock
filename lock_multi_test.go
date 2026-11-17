package go_redislock

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Test_MultiLock_Stubs 联锁(MultiLock)系列当前是未实现的存根,返回 ErrNotImplemented。
// 本测试固化"未实现返回哨兵错误"的现状并覆盖存根代码行,
// 后续真正实现联锁后,应将其替换为基于 mock 的行为测试。
func Test_MultiLock_Stubs(t *testing.T) {
	lock := New(nil, "unit_key").(*RedisLock)
	ctx := context.Background()

	cases := []struct {
		name    string
		call    func() error
		wantErr error
	}{
		{"MultiLock", func() error { return lock.MultiLock(ctx, nil) }, ErrNotImplemented},
		{"MultiUnLock", func() error { return lock.MultiUnLock(ctx, nil) }, ErrNotImplemented},
		{"SpinMultiLock", func() error { return lock.SpinMultiLock(ctx, nil, time.Second) }, ErrNotImplemented},
		{"MultiRenew", func() error { return lock.MultiRenew(ctx, nil) }, ErrNotImplemented},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("%s: 期望 %v,实际 %v", tt.name, tt.wantErr, err)
			}
		})
	}
}

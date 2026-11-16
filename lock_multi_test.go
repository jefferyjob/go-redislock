package go_redislock

import (
	"context"
	"testing"
	"time"
)

// Test_MultiLock_Stubs 联锁(MultiLock)系列当前是未实现的 panic 存根。
// 本测试固化"未实现即 panic"的现状并覆盖存根代码行,
// 后续真正实现联锁后,应将其替换为基于 mock 的行为测试。
func Test_MultiLock_Stubs(t *testing.T) {
	lock := New(nil, "unit_key").(*RedisLock)
	ctx := context.Background()

	cases := []struct {
		name string
		call func()
	}{
		{"MultiLock", func() { _ = lock.MultiLock(ctx, nil) }},
		{"MultiUnLock", func() { _ = lock.MultiUnLock(ctx, nil) }},
		{"SpinMultiLock", func() { _ = lock.SpinMultiLock(ctx, nil, time.Second) }},
		{"MultiRenew", func() { _ = lock.MultiRenew(ctx, nil) }},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s: 期望 panic,实际正常返回", tt.name)
				}
			}()
			tt.call()
		})
	}
}

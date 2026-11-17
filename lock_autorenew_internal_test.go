package go_redislock

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Test_RunRenewLoop_InvalidInterval 验证续期周期非正数时 runRenewLoop
// 直接返回:不执行任何续期,也不会因 NewTicker(0) panic。
// 该分支是防御性保护(正常入口 New() 已保证 lockTimeout 为正数)。
func Test_RunRenewLoop_InvalidInterval(t *testing.T) {
	var calls atomic.Int64
	done := make(chan struct{})

	go func() {
		defer close(done)
		runRenewLoop(context.Background(), 0, func(ctx context.Context) error {
			calls.Add(1)
			return nil
		})
	}()

	select {
	case <-done:
		// 正常返回
	case <-time.After(time.Second):
		t.Fatal("runRenewLoop 在 interval<=0 时应立即返回")
	}
	if calls.Load() != 0 {
		t.Errorf("不应执行任何续期,实际 %d 次", calls.Load())
	}
}

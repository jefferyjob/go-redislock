package go_redislock_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	go_redislock "github.com/jefferyjob/go-redislock"
	"github.com/jefferyjob/go-redislock/mocks"
)

// newCountingMock 返回一个按 ARGV 参数个数(1~3)路由、并统计调用次数的 mock:
// 首次调用(加锁)固定返回成功,后续调用(续期/解锁)由 renewResult 决定。
// 由于不同锁族的加锁/续期/解锁 ARGV 个数不同,这里按参数个数注册多个期望。
func newCountingMock(t *testing.T, calls *atomic.Int64, renewResult int64) *mocks.MockRedisInter {
	t.Helper()
	ctrl := gomock.NewController(t)
	mock := mocks.NewMockRedisInter(ctrl)

	fn := func(ctx context.Context, script string, keys []string, args ...interface{}) go_redislock.RedisCmd {
		if calls.Add(1) == 1 {
			return fakeCmd{result: 1}
		}
		return fakeCmd{result: renewResult}
	}

	for _, numArgs := range []int{1, 2, 3} {
		args := make([]interface{}, numArgs)
		for i := range args {
			args[i] = gomock.Any()
		}
		mock.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), args...).DoAndReturn(fn).AnyTimes()
	}
	return mock
}

// Test_AutoRenew 验证 WithAutoRenew 开启后的续期 goroutine 生命周期:
// 周期性续期、解锁后停止、续期失败后停止
func Test_AutoRenew(t *testing.T) {
	families := []struct {
		name   string
		lock   func(l go_redislock.RedisLockInter, ctx context.Context) error
		unlock func(l go_redislock.RedisLockInter, ctx context.Context) error
	}{
		{"Lock/UnLock",
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.Lock(ctx) },
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.UnLock(ctx) }},
		{"FairLock/FairUnLock",
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.FairLock(ctx, "req-1") },
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.FairUnLock(ctx, "req-1") }},
		{"RLock/RUnLock",
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.RLock(ctx) },
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.RUnLock(ctx) }},
		{"WLock/WUnLock",
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.WLock(ctx) },
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.WUnLock(ctx) }},
	}

	for _, f := range families {
		t.Run(f.name, func(t *testing.T) {
			t.Run("周期续期并在解锁后停止", func(t *testing.T) {
				var calls atomic.Int64
				mock := newCountingMock(t, &calls, 1)

				// 续期周期 = lockTimeout / 3 = 30ms
				lock := go_redislock.New(mock, "unit_key",
					go_redislock.WithToken("token-1"),
					go_redislock.WithAutoRenew(),
					go_redislock.WithTimeout(90*time.Millisecond))

				ctx := context.Background()
				if err := f.lock(lock, ctx); err != nil {
					t.Fatalf("加锁失败:%v", err)
				}
				base := calls.Load()

				// 等待出现至少 3 次续期
				waitFor(t, 2*time.Second, "等待周期续期发生", func() bool {
					return calls.Load()-base >= 3
				})

				// 解锁应取消续期 goroutine
				if err := f.unlock(lock, ctx); err != nil {
					t.Fatalf("解锁失败:%v", err)
				}
				after := calls.Load()

				// 留出多个续期周期的窗口,确认调用数不再增长
				// (允许 1 次取消瞬间恰好在途的续期调用)
				time.Sleep(150 * time.Millisecond)
				if got := calls.Load(); got > after+1 {
					t.Errorf("解锁后续期 goroutine 未停止:调用数 %d -> %d", after, got)
				}
			})

			t.Run("续期失败后停止", func(t *testing.T) {
				var calls atomic.Int64
				// 首次加锁成功,之后所有续期均失败
				mock := newCountingMock(t, &calls, 0)

				lock := go_redislock.New(mock, "unit_key",
					go_redislock.WithToken("token-1"),
					go_redislock.WithAutoRenew(),
					go_redislock.WithTimeout(90*time.Millisecond))

				ctx := context.Background()
				if err := f.lock(lock, ctx); err != nil {
					t.Fatalf("加锁失败:%v", err)
				}
				base := calls.Load()

				// 等待第一次续期(失败)发生
				waitFor(t, 2*time.Second, "等待首次续期发生", func() bool {
					return calls.Load()-base >= 1
				})

				// 续期失败后 goroutine 应退出,不再有新的续期调用
				time.Sleep(150 * time.Millisecond)
				if got := calls.Load(); got != base+1 {
					t.Errorf("续期失败后 goroutine 未停止:期望仅 1 次续期,实际 %d 次", got-base)
				}
			})
		})
	}
}

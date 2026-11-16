package go_redislock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	go_redislock "github.com/jefferyjob/go-redislock"
	"github.com/jefferyjob/go-redislock/mocks"
)

// Test_Spin 验证各锁族自旋变体的行为:
// 首次加锁成功、失败后重试成功、超时返回 ErrSpinLockTimeout、
// Context 取消返回 ErrSpinLockDone
func Test_Spin(t *testing.T) {
	families := []struct {
		name    string
		numArgs int
		spin    func(l go_redislock.RedisLockInter, ctx context.Context, timeout time.Duration) error
	}{
		{"SpinLock", 2, func(l go_redislock.RedisLockInter, ctx context.Context, timeout time.Duration) error {
			return l.SpinLock(ctx, timeout)
		}},
		{"SpinFairLock", 3, func(l go_redislock.RedisLockInter, ctx context.Context, timeout time.Duration) error {
			return l.SpinFairLock(ctx, "req-1", timeout)
		}},
		{"SpinRLock", 2, func(l go_redislock.RedisLockInter, ctx context.Context, timeout time.Duration) error {
			return l.SpinRLock(ctx, timeout)
		}},
		{"SpinWLock", 2, func(l go_redislock.RedisLockInter, ctx context.Context, timeout time.Duration) error {
			return l.SpinWLock(ctx, timeout)
		}},
	}

	for _, f := range families {
		t.Run(f.name, func(t *testing.T) {
			t.Run("首次加锁成功", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mock := mocks.NewMockRedisInter(ctrl)
				// 只允许恰好一次 Eval 调用,成功后不应继续自旋
				expectEval(mock, f.numArgs, fakeCmd{result: 1}).Times(1)

				lock := go_redislock.New(mock, "unit_key", go_redislock.WithToken("token-1"))
				if err := f.spin(lock, context.Background(), time.Second); err != nil {
					t.Fatalf("期望成功,实际 %v", err)
				}
			})

			t.Run("重试后成功", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mock := mocks.NewMockRedisInter(ctrl)
				// 第一次失败、100ms 后重试成功
				gomock.InOrder(
					expectEval(mock, f.numArgs, fakeCmd{result: 0}),
					expectEval(mock, f.numArgs, fakeCmd{result: 1}),
				)

				lock := go_redislock.New(mock, "unit_key", go_redislock.WithToken("token-1"))
				if err := f.spin(lock, context.Background(), time.Second); err != nil {
					t.Fatalf("期望重试后成功,实际 %v", err)
				}
			})

			t.Run("超时返回 ErrSpinLockTimeout", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mock := mocks.NewMockRedisInter(ctrl)
				expectEval(mock, f.numArgs, fakeCmd{result: 0}).AnyTimes()

				lock := go_redislock.New(mock, "unit_key", go_redislock.WithToken("token-1"))
				err := f.spin(lock, context.Background(), 50*time.Millisecond)
				if !errors.Is(err, go_redislock.ErrSpinLockTimeout) {
					t.Fatalf("期望 ErrSpinLockTimeout,实际 %v", err)
				}
			})

			t.Run("Context取消返回 ErrSpinLockDone", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mock := mocks.NewMockRedisInter(ctrl)
				expectEval(mock, f.numArgs, fakeCmd{result: 0}).AnyTimes()

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				lock := go_redislock.New(mock, "unit_key", go_redislock.WithToken("token-1"))
				err := f.spin(lock, ctx, time.Second)
				if !errors.Is(err, go_redislock.ErrSpinLockDone) {
					t.Fatalf("期望 ErrSpinLockDone,实际 %v", err)
				}
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("期望联合错误包含 context.Canceled,实际 %v", err)
				}
			})
		})
	}
}

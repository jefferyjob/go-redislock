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

// fakeCmd 是 RedisCmd 的固定实现,用于在单元测试中模拟 Lua 脚本的执行结果,
// 避免依赖真实 Redis。锁语义本身的正确性由 tests/ 模块的集成测试保障,
// 本包只验证 Go 层的行为:参数透传、返回值映射、自旋与自动续期等。
type fakeCmd struct {
	result int64
	err    error
}

// Result 模拟 Eval 的原始返回
func (f fakeCmd) Result() (interface{}, error) { return f.result, f.err }

// Int64 模拟 Eval 结果的整数转换
func (f fakeCmd) Int64() (int64, error) { return f.result, f.err }

// expectEval 声明一次对 Eval 的期望调用,numArgs 为 ARGV 参数个数
func expectEval(m *mocks.MockRedisInter, numArgs int, cmd fakeCmd) *gomock.Call {
	args := make([]interface{}, numArgs)
	for i := range args {
		args[i] = gomock.Any()
	}
	return m.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), args...).Return(cmd)
}

// waitFor 轮询等待条件满足,超时则失败
func waitFor(t *testing.T, timeout time.Duration, msg string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("等待超时(%v):%s", timeout, msg)
}

// Test_New_Options 验证构造选项是否正确透传给 Lua 脚本参数
func Test_New_Options(t *testing.T) {
	t.Run("默认 Token 与默认超时", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockRedisInter(ctrl)
		// 未设置 WithToken 时应生成唯一 token,未设置 WithTimeout 时默认 5s
		mock.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Eq([]string{"unit_key"}),
			gomock.Any(), int64(5000)).Return(fakeCmd{result: 1})

		lock := go_redislock.New(mock, "unit_key")
		if err := lock.Lock(context.Background()); err != nil {
			t.Fatalf("Lock() error = %v", err)
		}
	})

	t.Run("WithToken 与 WithTimeout 透传", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockRedisInter(ctrl)
		mock.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Eq([]string{"unit_key"}),
			"token-1", int64(3000)).Return(fakeCmd{result: 1})

		lock := go_redislock.New(mock, "unit_key",
			go_redislock.WithToken("token-1"),
			go_redislock.WithTimeout(3*time.Second))
		if err := lock.Lock(context.Background()); err != nil {
			t.Fatalf("Lock() error = %v", err)
		}
	})

	t.Run("WithRequestTimeout 透传给公平锁", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mock := mocks.NewMockRedisInter(ctrl)
		mock.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Eq([]string{"unit_key"}),
			"req-1", int64(5000), int64(7000)).Return(fakeCmd{result: 1})

		lock := go_redislock.New(mock, "unit_key",
			go_redislock.WithRequestTimeout(7*time.Second))
		if err := lock.FairLock(context.Background(), "req-1"); err != nil {
			t.Fatalf("FairLock() error = %v", err)
		}
	})
}

// Test_Result_Mapping 验证各锁族方法对 Lua 脚本返回值的映射:
// 脚本返回 1 = 成功(nil)、返回 0 = 失败(哨兵错误)、
// Eval 执行异常 = 包装为 ErrException + 原始错误的联合错误
func Test_Result_Mapping(t *testing.T) {
	evalErr := errors.New("mock redis eval error")

	methods := []struct {
		name    string
		numArgs int // Eval 的 ARGV 参数个数
		failErr error
		call    func(l go_redislock.RedisLockInter, ctx context.Context) error
	}{
		{"Lock", 2, go_redislock.ErrLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.Lock(ctx) }},
		{"UnLock", 1, go_redislock.ErrUnLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.UnLock(ctx) }},
		{"Renew", 2, go_redislock.ErrLockRenewFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.Renew(ctx) }},
		{"FairLock", 3, go_redislock.ErrLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.FairLock(ctx, "req-1") }},
		{"FairUnLock", 1, go_redislock.ErrUnLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.FairUnLock(ctx, "req-1") }},
		{"FairRenew", 2, go_redislock.ErrLockRenewFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.FairRenew(ctx, "req-1") }},
		{"RLock", 2, go_redislock.ErrLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.RLock(ctx) }},
		{"RUnLock", 1, go_redislock.ErrUnLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.RUnLock(ctx) }},
		{"RRenew", 2, go_redislock.ErrLockRenewFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.RRenew(ctx) }},
		{"WLock", 2, go_redislock.ErrLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.WLock(ctx) }},
		{"WUnLock", 1, go_redislock.ErrUnLockFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.WUnLock(ctx) }},
		{"WRenew", 2, go_redislock.ErrLockRenewFailed,
			func(l go_redislock.RedisLockInter, ctx context.Context) error { return l.WRenew(ctx) }},
	}

	const (
		outcomeSuccess   = 0
		outcomeSentinel  = 1
		outcomeException = 2
	)
	outcomes := []struct {
		name string
		cmd  fakeCmd
		kind int
	}{
		{"脚本返回1-成功", fakeCmd{result: 1}, outcomeSuccess},
		{"脚本返回0-失败", fakeCmd{result: 0}, outcomeSentinel},
		{"脚本执行异常", fakeCmd{result: 0, err: evalErr}, outcomeException},
	}

	for _, m := range methods {
		for _, o := range outcomes {
			name := m.name + "-" + o.name
			t.Run(name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mock := mocks.NewMockRedisInter(ctrl)
				expectEval(mock, m.numArgs, o.cmd)

				lock := go_redislock.New(mock, "unit_key", go_redislock.WithToken("token-1"))
				err := m.call(lock, context.Background())

				switch o.kind {
				case outcomeSuccess:
					if err != nil {
						t.Errorf("%s: 期望成功,实际 %v", name, err)
					}
				case outcomeSentinel:
					if !errors.Is(err, m.failErr) {
						t.Errorf("%s: 期望 %v,实际 %v", name, m.failErr, err)
					}
				case outcomeException:
					if !errors.Is(err, go_redislock.ErrException) || !errors.Is(err, evalErr) {
						t.Errorf("%s: 期望包装 ErrException 与原始错误,实际 %v", name, err)
					}
				}
			})
		}
	}
}

package go_redislock

import (
	"context"
	"errors"
	"github.com/go-redis/redismock/v9"
	"testing"
	"time"
)

func TestFairLock(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputReqId string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "Lua脚本执行成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", lockTime.Milliseconds(), lockTime.Milliseconds()).
					SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "test_key",
			inputReqId: "test_req_id",
			wantErr:    nil,
		},
		{
			name: "触发自动续期",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"key"},
					"req_id", lockTime.Milliseconds(), lockTime.Milliseconds()).
					SetVal(int64(1))
				mock.ExpectEval(fairRenewScript, []string{"key"},
					"req_id", lockTime.Milliseconds()).
					SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "key",
			inputReqId: "req_id",
			sleepTime:  5 * time.Second,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey,
				WithAutoRenew(),
				WithRequestTimeout(lockTime), // 设置公平锁请求超时时间
			)
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.sleepTime != time.Duration(0) {
				// 模拟业务执行时间
				time.Sleep(tt.sleepTime)
			}
		})
	}
}

func TestFairLockFairRenew(t *testing.T) {
	tests := []struct {
		name         string
		mock         func(t *testing.T, ctx context.Context) RedisInter
		inputKey     string
		inputReqId   string
		sleepTime    time.Duration // 模拟业务执行时间
		wantErr      error
		wantRenewErr error
	}{
		{
			name: "续期失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"key"},
					"req_id", lockTime.Milliseconds(), lockTime.Milliseconds()).
					SetVal(int64(1))
				mock.ExpectEval(fairRenewScript, []string{"key"},
					"req_id", lockTime.Milliseconds()).
					SetVal(int64(0))
				return NewRedisMockAdapter(db)
			},
			inputKey:     "key",
			inputReqId:   "req_id",
			sleepTime:    2 * time.Second,
			wantErr:      nil,
			wantRenewErr: ErrLockRenewFailed, // 续期失败
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey)
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			err = lock.FairRenew(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantRenewErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFairUnLock(t *testing.T) {
	tests := []struct {
		name          string
		mock          func(t *testing.T, ctx context.Context) RedisInter
		inputKey      string
		inputReqId    string
		wantLockErr   error
		wantUnlockErr error
	}{
		{
			name: "加锁成功后解锁成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			wantLockErr:   nil,
			wantUnlockErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey, WithAutoRenew())
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantLockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantLockErr)
			}

			err = lock.FairUnLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantUnlockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantUnlockErr)
			}
		})
	}
}

// 公平锁错误测试
func TestFairLockErr(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputReqId string
		wantErr    error
	}{
		{
			name: "Lua脚本执行异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).
					SetErr(ErrException)
				return NewRedisMockAdapter(db)
			},
			inputKey:   "test_key",
			inputReqId: "test_req_id",
			wantErr:    ErrException,
		},
		{
			name: "公平锁-获取锁资源失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).
					SetVal(int64(0))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "test_key",
			inputReqId: "test_req_id",
			wantErr:    ErrLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey)
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 公平锁解锁错误测试
func TestFairUnLockErr(t *testing.T) {
	tests := []struct {
		name          string
		mock          func(t *testing.T, ctx context.Context) RedisInter
		inputKey      string
		inputReqId    string
		wantLockErr   error
		wantUnlockErr error
	}{
		{
			name: "公平锁-解锁-Lua脚本执行异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").SetErr(ErrException)
				return NewRedisMockAdapter(db)
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			wantLockErr:   nil,
			wantUnlockErr: ErrException,
		},
		{
			name: "公平锁-解锁-解锁失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).
					SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").
					SetVal(int64(0))
				return NewRedisMockAdapter(db)
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			wantLockErr:   nil,
			wantUnlockErr: ErrUnLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey)
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantLockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantLockErr)
			}

			err = lock.FairUnLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantUnlockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantUnlockErr)
			}
		})
	}
}

func TestSpinFairLock(t *testing.T) {
	tests := []struct {
		name        string
		mock        func(t *testing.T, ctx context.Context) RedisInter
		before      func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		after       func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		inputKey    string
		inputReqId  string
		spinTimeout time.Duration
		wantErr     error
	}{
		{
			name: "自旋公平锁-成功获取锁",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			inputKey:    "test_key",
			inputReqId:  "test_req_id",
			spinTimeout: 5 * time.Second,
			wantErr:     nil,
		},
		{
			name: "自旋公平锁-超时",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 3秒 = 30 * 100毫秒
				for i := 0; i < 30; i++ {
					mock.ExpectEval(fairLockScript, []string{"test_key"},
						"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).SetVal(int64(0))
				}
				return NewRedisMockAdapter(db)
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			inputKey:    "test_key",
			inputReqId:  "test_req_id",
			spinTimeout: 2 * time.Second,
			wantErr:     ErrSpinLockTimeout,
		},
		{
			name: "自旋公平锁 Context Cancel",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 5秒 = 50 * 100毫秒
				for i := 0; i < 50; i++ {
					mock.ExpectEval(fairLockScript, []string{"test_key"},
						"test_req_id", 5*time.Second.Milliseconds(), 5*time.Second.Milliseconds()).SetVal(int64(0))
				}
				return NewRedisMockAdapter(db)
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				go func() {
					time.Sleep(2 * time.Second)
					cancel()
				}()
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			inputKey:    "test_key",
			inputReqId:  "test_req_id",
			spinTimeout: 5 * time.Second,
			wantErr:     ErrSpinLockDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			lock := New(tt.mock(t, ctx), tt.inputKey, WithRequestTimeout(5*time.Second))

			tt.before(t, ctx, cancel, lock)
			err := lock.SpinFairLock(ctx, tt.inputReqId, tt.spinTimeout)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, got = %v", tt.wantErr, err)
			}
			tt.after(t, ctx, cancel, lock)
		})
	}
}

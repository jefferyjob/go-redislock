package go_redislock

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisLock_Lock(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(t *testing.T) *redis.Client
		inputKey   string
		inputToken string
		wantErr    error
	}{
		{
			name: "加锁成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				return db
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "加锁内部异常",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetErr(ErrException)
				return db
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    ErrException,
		},
		{
			name: "加锁失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("nil")
				return db
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    ErrLockFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lock := New(context.TODO(), tc.mock(t), tc.inputKey, WithToken(tc.inputToken))
			err := lock.Lock()
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestRedisLock_UnLock(t *testing.T) {
	testCases := []struct {
		name       string
		before     func(t *testing.T, lock RedisLockInter)
		after      func(t *testing.T, lock RedisLockInter)
		mock       func(t *testing.T) *redis.Client
		inputKey   string
		inputToken string
		wantErr    error
	}{
		{
			name: "解锁成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				mock.ExpectEval(unLockScript, []string{"key"}, "token").
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
				err := lock.Lock()
				require.NoError(t, err)
			},
			after: func(t *testing.T, lock RedisLockInter) {
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "解锁内部异常",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				mock.ExpectEval(unLockScript, []string{"key"}, "token").
					SetErr(ErrException)
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
				err := lock.Lock()
				require.NoError(t, err)
			},
			after: func(t *testing.T, lock RedisLockInter) {
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    ErrException,
		},
		{
			name: "解锁失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				mock.ExpectEval(unLockScript, []string{"key"}, "token").
					SetVal("nil")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
				err := lock.Lock()
				require.NoError(t, err)
			},
			after: func(t *testing.T, lock RedisLockInter) {
			},
			inputKey:   "key",
			inputToken: "token",
			wantErr:    ErrUnLockFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lock := New(context.TODO(), tc.mock(t), tc.inputKey, WithToken(tc.inputToken))

			tc.before(t, lock)

			err := lock.UnLock()
			assert.Equal(t, tc.wantErr, err)

			tc.after(t, lock)
		})
	}
}

func TestRedisLock_SpinLock(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(t *testing.T) *redis.Client
		before       func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		after        func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		inputKey     string
		inputToken   string
		inputTimeout time.Duration
		wantErr      error
	}{
		{
			name: "自旋锁-加锁成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {

			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {

			},
			inputTimeout: time.Second * 2,
			inputKey:     "key",
			inputToken:   "token",
			wantErr:      nil,
		},
		{
			name: "自旋锁-加锁超时",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				go func() {
					err := lock.Lock()
					require.NoError(t, err)
					defer lock.UnLock()
					time.Sleep(time.Second * 6) // 模拟业务执行实行
				}()
				time.Sleep(time.Second * 1) // 确保协程加锁成功
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {

			},
			inputTimeout: time.Second * 3,
			inputKey:     "key",
			inputToken:   "token",
			wantErr:      ErrSpinLockTimeout,
		},
		{
			name: "自旋锁-Ctx取消",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(lockScript, []string{"key"}, "token", lockTime.Seconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				go func() {
					err := lock.Lock()
					require.NoError(t, err)
					defer lock.UnLock()
					time.Sleep(time.Second * 6) // 模拟业务执行实行
				}()
				go func() {
					time.Sleep(time.Second * 2)
					cancel()
				}()
				time.Sleep(time.Second * 1) // 确保协程加锁成功
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {

			},
			inputTimeout: time.Second * 10,
			inputKey:     "key",
			inputToken:   "token",
			wantErr:      ErrSpinLockDone,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			lock := New(ctx, tc.mock(t), tc.inputKey, WithToken(tc.inputToken))

			tc.before(t, ctx, cancel, lock)

			err := lock.SpinLock(tc.inputTimeout)
			assert.Equal(t, tc.wantErr, err)

			tc.after(t, ctx, cancel, lock)
		})
	}
}

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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
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

func TestRedisLock_LockRenew(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(t *testing.T) *redis.Client
		before     func(t *testing.T, lock RedisLockInter)
		after      func(t *testing.T, lock RedisLockInter)
		inputKey   string
		inputToken string
		inputSleep time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "锁手动续期成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantRenewScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
			},
			after: func(t *testing.T, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:   "key",
			inputToken: "token",
			inputSleep: time.Second * 10,
			wantErr:    nil,
		},
		{
			name: "锁手动续期异常",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantRenewScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetErr(ErrException)
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
			},
			after: func(t *testing.T, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:   "key",
			inputToken: "token",
			inputSleep: time.Second * 10,
			wantErr:    ErrException,
		},
		{
			name: "锁手动续期失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantRenewScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("nil")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
			},
			after: func(t *testing.T, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:   "key",
			inputToken: "token",
			inputSleep: time.Second * 10,
			wantErr:    ErrLockRenewFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lock := New(context.TODO(), tc.mock(t), tc.inputKey, WithToken(tc.inputToken))
			tc.before(t, lock)

			err := lock.Lock()
			require.NoError(t, err)

			// 第6秒，手动续期
			go func() {
				time.Sleep(time.Second * 6)
				errRenew := lock.Renew()
				assert.Equal(t, tc.wantErr, errRenew)
			}()

			time.Sleep(tc.inputSleep) // 模拟业务执行时间

			tc.after(t, lock)
		})
	}
}

func TestRedisLock_LockAutoRenew(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(t *testing.T) *redis.Client
		before     func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		after      func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter)
		inputKey   string
		inputToken string
		inputSleep time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "锁自动续期成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantRenewScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:   "key",
			inputToken: "token",
			inputSleep: time.Second * 10,
			wantErr:    nil,
		},
		{
			name: "锁自动续期-Ctx取消",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				mock.ExpectEval(reentrantRenewScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				go func() {
					time.Sleep(time.Second * 3)
					cancel()
				}()
			},
			after: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:   "key",
			inputToken: "token",
			inputSleep: time.Second * 10,
			wantErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			lock := New(ctx, tc.mock(t), tc.inputKey,
				WithToken(tc.inputToken),
				WithAutoRenew(),
			)
			tc.before(t, ctx, cancel, lock)

			err := lock.Lock()
			require.NoError(t, err)
			time.Sleep(tc.inputSleep) // 模拟业务执行时间

			tc.after(t, ctx, cancel, lock)
		})
	}
}

func TestRedisLock_LockTimeout(t *testing.T) {
	testCases := []struct {
		name             string
		before           func(t *testing.T, lock RedisLockInter)
		after            func(t *testing.T, lock RedisLockInter)
		mock             func(t *testing.T) *redis.Client
		inputWithTimeout time.Duration
		inputKey         string
		inputToken       string
		inputSleep       time.Duration
		wantErr          error
	}{
		{
			name: "加锁成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				// 第一次加锁
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", (time.Second * 2).Milliseconds()).
					SetVal("OK")
				// 第二次加锁
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", (time.Second * 2).Milliseconds()).
					SetVal("OK")

				// 第一次解锁
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
					SetVal("OK")
				// 第二次解锁
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
				go func() {
					err := lock.Lock()
					require.NoError(t, err)
					defer lock.UnLock()
					time.Sleep(time.Second * 2)
				}()
			},
			after: func(t *testing.T, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:         "key",
			inputToken:       "token",
			inputWithTimeout: time.Second * 2,
			inputSleep:       time.Second * 4,
			wantErr:          nil,
		},
		{
			name: "加锁失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				// 第一次加锁
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", (time.Second * 5).Milliseconds()).
					SetVal("OK")
				// 第二次加锁
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", (time.Second * 5).Milliseconds()).
					SetVal("nil")

				// 第一次解锁
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
					SetVal("OK")
				// 第二次解锁
				mock.ExpectEval(reentrantUnLockScript, []string{"key"}, "token").
					SetVal("OK")
				return db
			},
			before: func(t *testing.T, lock RedisLockInter) {
				go func() {
					err := lock.Lock()
					require.NoError(t, err)
					defer lock.UnLock()
					time.Sleep(time.Second * 5)
				}()
			},
			after: func(t *testing.T, lock RedisLockInter) {
				_ = lock.UnLock()
			},
			inputKey:         "key",
			inputToken:       "token",
			inputWithTimeout: time.Second * 5,
			inputSleep:       time.Second * 3,
			wantErr:          ErrLockFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lock := New(context.TODO(), tc.mock(t), tc.inputKey,
				WithToken(tc.inputToken),
				WithTimeout(tc.inputWithTimeout),
			)

			tc.before(t, lock)

			time.Sleep(tc.inputSleep)

			err := lock.Lock()
			assert.Equal(t, tc.wantErr, err)

			tc.after(t, lock)
		})
	}
}

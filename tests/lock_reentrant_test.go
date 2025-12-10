package tests

//
// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"
//
// 	redislock "github.com/jefferyjob/go-redislock"
// 	"github.com/stretchr/testify/require"
// )
//
// func Test_Lock(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name       string
// 		inputKey   string
// 		inputToken string
// 		before     func(inputKey string)
// 		wantErr    error
// 	}{
// 		{
// 			name:       "创建锁对象-正常",
// 			inputKey:   "test_key",
// 			inputToken: "test_token",
// 			wantErr:    nil,
// 		},
// 		{
// 			name:       "创建锁对象-失败",
// 			inputKey:   "test_key",
// 			inputToken: "test_token",
// 			before: func(inputKey string) {
// 				ctx := context.Background()
// 				lock := redislock.New(adapter, inputKey, redislock.WithToken("other_token"))
// 				_ = lock.Lock(ctx)
// 			},
// 			wantErr: redislock.ErrLockFailed,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.before != nil {
// 				tt.before(tt.inputKey)
// 			}
//
// 			ctx := context.Background()
// 			lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputToken))
// 			err := lock.Lock(ctx)
// 			if !errors.Is(err, tt.wantErr) {
// 				t.Errorf("expected error %v, got %v", tt.wantErr, err)
// 			}
// 			defer lock.UnLock(ctx)
// 		})
// 	}
// }
//
// func Test_UnLock(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name       string
// 		inputKey   string
// 		inputToken string
// 		between    func(inputKey string, inputToken string)
// 		wantErr    error
// 	}{
// 		{
// 			name:       "释放锁对象-正常",
// 			inputKey:   "test_key",
// 			inputToken: "test_token",
// 			wantErr:    nil,
// 		},
// 		{
// 			name:       "创建锁对象-失败",
// 			inputKey:   "test_key",
// 			inputToken: "test_token",
// 			between: func(inputKey string, inputToken string) {
// 				ctx := context.Background()
// 				lock := redislock.New(adapter, inputKey, redislock.WithToken(inputToken))
// 				_ = lock.UnLock(ctx)
// 			},
// 			wantErr: redislock.ErrUnLockFailed,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputToken))
// 			err := lock.Lock(ctx)
// 			if err != nil {
// 				require.NoError(t, err)
// 			}
//
// 			if tt.between != nil {
// 				tt.between(tt.inputKey, tt.inputToken)
// 			}
//
// 			err = lock.UnLock(ctx)
// 			if !errors.Is(err, tt.wantErr) {
// 				t.Errorf("expected error %v, got %v", tt.wantErr, err)
// 			}
// 		})
// 	}
// }
//
// func Test_SpinLock(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name         string
// 		inputKey     string
// 		inputToken   string
// 		before       func(ctx context.Context, lock redislock.RedisLockInter, cancel context.CancelFunc)
// 		inputTimeout time.Duration
// 		wantErr      error
// 	}{
// 		{
// 			name:       "自旋锁-成功加锁",
// 			inputKey:   "spin_key_success",
// 			inputToken: "token_success",
// 			before: func(ctx context.Context, lock redislock.RedisLockInter, cancel context.CancelFunc) {
// 				// 无操作，让 lock.SpinLock 能直接成功
// 			},
// 			inputTimeout: time.Second * 3,
// 			wantErr:      nil,
// 		},
// 		{
// 			name:       "自旋锁-超时未拿到锁",
// 			inputKey:   "spin_key_timeout",
// 			inputToken: "token_timeout",
// 			before: func(ctx context.Context, lock redislock.RedisLockInter, cancel context.CancelFunc) {
// 				// 提前上锁，阻塞后续尝试
// 				go func() {
// 					owner := redislock.New(adapter, "spin_key_timeout", redislock.WithToken("other"))
// 					_ = owner.Lock(ctx)
// 					time.Sleep(time.Second * 5) // 保持锁
// 					_ = owner.UnLock(ctx)
// 				}()
// 				time.Sleep(time.Millisecond * 500)
// 			},
// 			inputTimeout: time.Second * 2,
// 			wantErr:      redislock.ErrSpinLockTimeout,
// 		},
// 		{
// 			name:       "自旋锁-Ctx取消",
// 			inputKey:   "spin_key_cancel",
// 			inputToken: "token_cancel",
// 			before: func(ctx context.Context, lock redislock.RedisLockInter, cancel context.CancelFunc) {
// 				go func() {
// 					owner := redislock.New(adapter, "spin_key_cancel", redislock.WithToken("other"))
// 					_ = owner.Lock(ctx)
// 					time.Sleep(time.Second * 5)
// 					_ = owner.UnLock(ctx)
// 				}()
// 				time.Sleep(time.Millisecond * 500)
// 				go func() {
// 					time.Sleep(time.Second)
// 					cancel() // 主动取消 context
// 				}()
// 			},
// 			inputTimeout: time.Second * 5,
// 			wantErr:      redislock.ErrSpinLockDone,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithCancel(context.Background())
// 			lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputToken))
//
// 			if tt.before != nil {
// 				tt.before(ctx, lock, cancel)
// 			}
//
// 			err := lock.SpinLock(ctx, tt.inputTimeout)
// 			if !errors.Is(err, tt.wantErr) {
// 				t.Errorf("expected error %v, got %v", tt.wantErr, err)
// 			}
// 			defer lock.UnLock(ctx)
// 		})
// 	}
// }
//
// func Test_LockRenew(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name            string
// 		inputKey        string
// 		inputToken      string
// 		inputRenewToken string
// 		inputSleep      time.Duration
// 		wantErr         error
// 	}{
// 		{
// 			name:            "锁手动续期成功",
// 			inputKey:        "test_key",
// 			inputToken:      "token_ok",
// 			inputRenewToken: "token_ok",
// 			inputSleep:      6 * time.Second,
// 			wantErr:         nil,
// 		},
// 		{
// 			name:            "锁手动续期失败-token不匹配",
// 			inputKey:        "test_key",
// 			inputToken:      "token_fail",
// 			inputRenewToken: "token_other",
// 			inputSleep:      6 * time.Second,
// 			wantErr:         redislock.ErrLockRenewFailed,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputToken))
// 			err := lock.Lock(ctx)
// 			require.NoError(t, err)
// 			defer lock.UnLock(ctx)
//
// 			go func() {
// 				time.Sleep(time.Second * 2)
// 				ctx := context.Background()
// 				lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputRenewToken))
// 				err = lock.Renew(ctx)
// 				if !errors.Is(err, tt.wantErr) {
// 					t.Errorf("expected error %v, got %v", tt.wantErr, err)
// 				}
// 			}()
//
// 			// 等待一段时间，确保续期操作完成
// 			time.Sleep(tt.inputSleep)
// 		})
// 	}
// }
//
// func Test_LockAutoRenew(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name       string
// 		inputKey   string
// 		inputToken string
// 		inputSleep time.Duration
// 		cancelTime time.Duration
// 	}{
// 		{
// 			name:       "锁自动续期成功",
// 			inputKey:   "test_key_auto_ok",
// 			inputToken: "token_auto_ok",
// 			inputSleep: 10 * time.Second,
// 			cancelTime: 0, // 不提前取消
// 		},
// 		{
// 			name:       "锁自动续期-提前取消ctx",
// 			inputKey:   "test_key_auto_cancel",
// 			inputToken: "token_auto_cancel",
// 			inputSleep: 10 * time.Second,
// 			cancelTime: 3 * time.Second,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithCancel(context.Background())
// 			lock := redislock.New(adapter, tt.inputKey,
// 				redislock.WithToken(tt.inputToken),
// 				redislock.WithAutoRenew(),
// 			)
//
// 			err := lock.Lock(ctx)
// 			require.NoError(t, err)
//
// 			if tt.cancelTime > 0 {
// 				go func() {
// 					time.Sleep(tt.cancelTime)
// 					cancel()
// 				}()
// 			}
//
// 			time.Sleep(tt.inputSleep)
// 			_ = lock.UnLock(context.Background())
// 		})
// 	}
// }

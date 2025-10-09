package go_redislock

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestWLock(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "写锁-加锁成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "写锁-加锁失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(0))
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrLockFailed,
		},
		{
			name: "写锁-加锁异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetErr(ErrException)
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey, WithToken(tt.inputToken))
			err := lock.WLock(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			// 释放锁
			err = lock.WUnLock(ctx)
			if err != nil {
				t.Errorf("Failed to unlock: %v", err)
			}
		})
	}
}

func TestWUnLock(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "写锁-解锁成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "写锁-解锁失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").SetVal(int64(0))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrUnLockFailed,
		},
		{
			name: "写锁-解锁异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(writeUnLockScript, []string{"testKey"}, "token").
					SetErr(ErrException)
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := New(tt.mock(t, ctx), tt.inputKey, WithToken(tt.inputToken))
			err := lock.WLock(ctx)
			if err != nil {
				t.Errorf("Failed to lock: %v", err)
			}

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			// 释放锁
			err = lock.WUnLock(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 测试读锁升级为写锁
func TestWLockByReadLock(t *testing.T) {
	ctx := context.Background()

	var (
		wg = &sync.WaitGroup{}
	)

	// mock Redis
	db, mock := redismock.NewClientMock()
	mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).SetVal(int64(1))  // 读锁-加锁成功
	mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).SetVal(int64(1)) // 写锁-加锁成功
	adapter := NewRedisMockAdapter(db)

	f1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		lock := New(adapter, "testKey", WithToken("token"))
		err := lock.RLock(ctx)
		time.Sleep(time.Second * 2) // 确保读锁保持执行，留给写锁足够的事情抢夺
		if err != nil {
			t.Errorf("Failed to lock: %v", err)
		}
	}

	f2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(time.Second * 1) // 确保读锁已经加锁成功
		lock := New(adapter, "testKey", WithToken("token"))
		err := lock.WLock(ctx)
		if err != nil {
			t.Errorf("Failed to lock: %v", err)
		}
	}

	wg.Add(2)
	go f1(wg)
	go f2(wg)
	wg.Wait()
}

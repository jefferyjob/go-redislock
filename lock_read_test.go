package go_redislock

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestRLock(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "读锁-加锁成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "读锁-加锁失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(0))
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrLockFailed,
		},
		{
			name: "读锁-加锁异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetErr(ErrException)
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
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
			err := lock.RLock(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			// 释放锁
			err = lock.RUnLock(ctx)
			if err != nil {
				t.Errorf("Failed to unlock: %v", err)
			}
		})
	}
}

func TestRUnLock(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name: "读锁-解锁成功",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").SetVal(int64(1))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    nil,
		},
		{
			name: "读锁-解锁失败",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").SetVal(int64(0))
				return NewRedisMockAdapter(db)
			},
			inputKey:   "testKey",
			inputToken: "token",
			wantErr:    ErrUnLockFailed,
		},
		{
			name: "读锁-解锁异常",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := redismock.NewClientMock()
				// 加锁
				mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				// 解锁
				mock.ExpectEval(readUnLockScript, []string{"testKey"}, "token").
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
			err := lock.RLock(ctx)
			if err != nil {
				t.Errorf("Failed to lock: %v", err)
			}

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			// 释放锁
			err = lock.RUnLock(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 测试读锁和写锁的互斥性
func TestRLockByWriteLock(t *testing.T) {
	ctx := context.Background()

	var (
		wg = &sync.WaitGroup{}
	)

	// mock Redis
	db, mock := redismock.NewClientMock()
	mock.ExpectEval(writeLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).SetVal(int64(1)) // 写锁-加锁成功
	mock.ExpectEval(readLockScript, []string{"testKey"}, "token", lockTime.Milliseconds()).SetVal(int64(0))  // 读锁-加锁成功
	adapter := NewRedisMockAdapter(db)

	f1 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		lock := New(adapter, "testKey", WithToken("token"))
		err := lock.WLock(ctx)
		if err != nil {
			t.Errorf("Failed to lock: %v", err)
		}
	}

	f2 := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(1 * time.Second) // 确保写锁已经加锁成功
		lock := New(adapter, "testKey", WithToken("token"))
		err := lock.RLock(ctx)
		if err == nil {
			t.Errorf("Failed to lock: %v", err)
		}
	}

	wg.Add(2)
	go f1(wg)
	go f2(wg)
	wg.Wait()
}

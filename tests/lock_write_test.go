package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
)

// 测试读锁升级为写锁
func Test_WLockByReadLock(t *testing.T) {
	adapter := getRedisClient()

	tests := []struct {
		name        string
		inputKey    string
		inputRToken string
		inputWToken string
		wantRErr    error
		wantWErr    error
	}{
		{
			name:        "读锁和写锁同token，读锁升级写锁成功",
			inputKey:    "testKey",
			inputRToken: "testToken",
			inputWToken: "testToken",
		},
		{
			name:        "读锁和写锁不同token，读锁升级写锁失败",
			inputKey:    "testKey",
			inputRToken: "testRToken",
			inputWToken: "testWToken",
			wantWErr:    redislock.ErrLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				wg = &sync.WaitGroup{}
			)

			// 线程1：读锁
			f1 := func(wg *sync.WaitGroup) {
				defer wg.Done()
				ctx := context.Background()
				lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputRToken))
				err := lock.RLock(ctx)
				if !errors.Is(err, tt.wantRErr) {
					t.Errorf("Failed to Rlock: %v", err)
				}
				defer lock.RUnLock(ctx)
				time.Sleep(time.Second * 2) // 确保读锁保持执行，留给写锁足够的事情抢夺

			}

			// 线程2：写锁
			f2 := func(wg *sync.WaitGroup) {
				defer wg.Done()
				time.Sleep(time.Second * 1) // 确保读锁已经加锁成功
				ctx := context.Background()
				lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputWToken))
				err := lock.WLock(ctx)
				defer lock.WUnLock(ctx)
				if !errors.Is(err, tt.wantWErr) {
					t.Errorf("Failed to Wlock: %v", err)
				}
			}

			wg.Add(2)
			go f1(wg)
			go f2(wg)
			wg.Wait()
		})
	}
}

// 测试读锁可重入
func TestWLockReentrant(t *testing.T) {
	adapter := getRedisClient()

	var (
		inputKey   = "testKey"
		inputToken = "testToken"
	)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		lock := redislock.New(adapter, inputKey,
			redislock.WithToken(inputToken),
			redislock.WithAutoRenew(),
		)
		err := lock.WLock(ctx)
		if err != nil {
			t.Errorf("Failed to lock: %v", err)
		}

		time.Sleep(time.Second * 1)

		if i == 4 {
			lock.WUnLock(ctx)
		}
	}
}

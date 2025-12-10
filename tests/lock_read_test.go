package tests

//
// import (
// 	"context"
// 	"errors"
// 	"sync"
// 	"testing"
// 	"time"
//
// 	redislock "github.com/jefferyjob/go-redislock"
// )
//
// // 测试写锁抢夺读锁
// func TestRLockByWriteLock(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	tests := []struct {
// 		name        string
// 		inputKey    string
// 		inputRToken string
// 		inputWToken string
// 		wantRErr    error
// 		wantWErr    error
// 	}{
// 		{
// 			name:        "读锁抢夺写锁，自己持有写锁：允许同时持有读锁，成功",
// 			inputKey:    "testKey",
// 			inputRToken: "testToken",
// 			inputWToken: "testToken",
// 		},
// 		{
// 			name:        "读锁抢夺写锁，他人持有写锁，失败",
// 			inputKey:    "testKey2",
// 			inputRToken: "testRToken",
// 			inputWToken: "testWToken",
// 			wantRErr:    redislock.ErrLockFailed,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var (
// 				wg = &sync.WaitGroup{}
// 			)
//
// 			// 线程1：写锁
// 			f1 := func(wg *sync.WaitGroup) {
// 				defer wg.Done()
// 				ctx := context.Background()
// 				lock := redislock.New(adapter, tt.inputKey,
// 					redislock.WithToken(tt.inputWToken),
// 					// redislock.WithTimeout(time.Second*60),
// 				)
// 				err := lock.WLock(ctx)
// 				if !errors.Is(err, tt.wantWErr) {
// 					t.Errorf("Failed to Wlock: %v", err)
// 				}
// 				time.Sleep(time.Second * 5) // 确保写锁保持执行，留给读锁足够的事情抢夺
// 				defer lock.WUnLock(ctx)
// 			}
//
// 			// 线程2：读锁
// 			f2 := func(wg *sync.WaitGroup) {
// 				defer wg.Done()
// 				time.Sleep(1 * time.Second) // 确保写锁已经加锁成功
// 				ctx := context.Background()
// 				lock := redislock.New(adapter, tt.inputKey,
// 					redislock.WithToken(tt.inputRToken),
// 					// redislock.WithTimeout(time.Second*60),
// 				)
// 				err := lock.RLock(ctx)
// 				if !errors.Is(err, tt.wantRErr) {
// 					t.Errorf("Failed to Rlock: %v", err)
// 				}
// 				defer lock.RUnLock(ctx)
// 			}
//
// 			wg.Add(2)
// 			go f1(wg)
// 			go f2(wg)
// 			wg.Wait()
// 		})
// 	}
// }
//
// // 读锁的可重入能力
// func TestRLockReentrant(t *testing.T) {
// 	adapter := getRedisClient()
//
// 	var (
// 		inputKey   = "testKey"
// 		inputToken = "testToken"
// 	)
//
// 	ctx := context.Background()
//
// 	for i := 0; i < 5; i++ {
// 		lock := redislock.New(adapter, inputKey,
// 			redislock.WithToken(inputToken),
// 			redislock.WithAutoRenew(),
// 		)
// 		err := lock.RLock(ctx)
// 		if err != nil {
// 			t.Errorf("Failed to lock: %v", err)
// 		}
// 		time.Sleep(time.Second * 1)
// 	}
// }

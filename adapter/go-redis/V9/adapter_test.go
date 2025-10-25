package v9

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
	"github.com/redis/go-redis/v9"
)

var (
	addr         = "127.0.0.1"
	port         = "63790"
	luaSetScript = `return redis.call("SET", KEYS[1], ARGV[1])`
	luaGetScript = `return redis.call("GET", KEYS[1])`
	luaDelScript = `return redis.call("DEL", KEYS[1])`
)

// Redis服务器测试
// 下面的代码将借助 redis 服务器进行测试，可以更加方便的测试服务中的问题
// 你可以实用下面的命令启动一个redis容器进行测试
// docker run -d -p 63790:6379 --name go_redis_lock redis
func getRedisClient() (redislock.RedisInter, *redis.Client) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	})
	rdbAdapter := New(rdb)
	return rdbAdapter, rdb
}

// 适配器测试
func TestAdapter(t *testing.T) {
	adapter, _ := getRedisClient()
	if adapter == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key"

	// 线程2抢占锁资源-预期失败
	go func() {
		time.Sleep(time.Second * 1)
		lock := redislock.New(adapter, key)
		err := lock.Lock(ctx)
		if err == nil {
			t.Errorf("Lock() returned unexpected success: %v", err)
			return
		}
		log.Println("线程2：抢占锁失败，锁已被其他线程占用")
	}()

	// 线程1加锁-预期成功
	lock := redislock.New(adapter, key)
	err := lock.Lock(ctx)
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock(ctx)

	// 模拟业务处理
	log.Println("线程1：锁已获取，开始执行任务")
	time.Sleep(time.Second * 5)
}

// 测试读锁升级为写锁
func TestWLockByReadLock(t *testing.T) {
	adapter, _ := getRedisClient()
	if adapter == nil {
		log.Println("Github actions skip this test")
		return
	}

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
				time.Sleep(time.Second * 2) // 确保读锁保持执行，留给写锁足够的事情抢夺
				defer lock.RUnLock(ctx)
			}

			// 线程2：写锁
			f2 := func(wg *sync.WaitGroup) {
				defer wg.Done()
				time.Sleep(time.Second * 1) // 确保读锁已经加锁成功
				ctx := context.Background()
				lock := redislock.New(adapter, tt.inputKey, redislock.WithToken(tt.inputWToken))
				err := lock.WLock(ctx)
				if !errors.Is(err, tt.wantWErr) {
					t.Errorf("Failed to Wlock: %v", err)
				}
				defer lock.WUnLock(ctx)
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
	adapter, _ := getRedisClient()
	if adapter == nil {
		log.Println("Github actions skip this test")
		return
	}

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

// 测试写锁抢夺读锁
func TestRLockByWriteLock(t *testing.T) {
	adapter, _ := getRedisClient()
	if adapter == nil {
		log.Println("Github actions skip this test")
		return
	}

	tests := []struct {
		name        string
		inputKey    string
		inputRToken string
		inputWToken string
		wantRErr    error
		wantWErr    error
	}{
		{
			name:        "读锁抢夺写锁，自己持有写锁：允许同时持有读锁，成功",
			inputKey:    "testKey",
			inputRToken: "testToken",
			inputWToken: "testToken",
		},
		{
			name:        "读锁抢夺写锁，他人持有写锁，失败",
			inputKey:    "testKey2",
			inputRToken: "testRToken",
			inputWToken: "testWToken",
			wantRErr:    redislock.ErrLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				wg = &sync.WaitGroup{}
			)

			// 线程1：写锁
			f1 := func(wg *sync.WaitGroup) {
				defer wg.Done()
				ctx := context.Background()
				lock := redislock.New(adapter, tt.inputKey,
					redislock.WithToken(tt.inputWToken),
					// redislock.WithTimeout(time.Second*60),
				)
				err := lock.WLock(ctx)
				if !errors.Is(err, tt.wantWErr) {
					t.Errorf("Failed to Wlock: %v", err)
				}
				time.Sleep(time.Second * 5) // 确保写锁保持执行，留给读锁足够的事情抢夺
				defer lock.WUnLock(ctx)
			}

			// 线程2：读锁
			f2 := func(wg *sync.WaitGroup) {
				defer wg.Done()
				time.Sleep(1 * time.Second) // 确保写锁已经加锁成功
				ctx := context.Background()
				lock := redislock.New(adapter, tt.inputKey,
					redislock.WithToken(tt.inputRToken),
					// redislock.WithTimeout(time.Second*60),
				)
				err := lock.RLock(ctx)
				if !errors.Is(err, tt.wantRErr) {
					t.Errorf("Failed to Rlock: %v", err)
				}
				defer lock.RUnLock(ctx)
			}

			wg.Add(2)
			go f1(wg)
			go f2(wg)
			wg.Wait()
		})
	}
}

// 读锁的可重入能力
func TestRLockReentrant(t *testing.T) {
	adapter, _ := getRedisClient()
	if adapter == nil {
		log.Println("Github actions skip this test")
		return
	}

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
		err := lock.RLock(ctx)
		if err != nil {
			t.Errorf("Failed to lock: %v", err)
		}
		time.Sleep(time.Second * 1)
	}
}

package go_redislock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// Redis服务器测试
// 下面的代码将借助 redis 服务器进行测试，可以更加方便的测试服务中的问题
// 你可以实用下面的命令启动一个redis容器进行测试
// docker run -d -p 6379:6379 --name go_redis_lock redis
// 注意：该服务在 GITHUB ACTIONS 并不会被测试
func getRedisClient() *redis.Client {
	return nil
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:63790",
	})

	// 尝试执行PING命令
	// 如果执行PING命令出错，则表明连接失败
	ctx := context.TODO()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	return rdb
}

func TestSevLock(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.TODO()
	key := "test_key_TestSevLock"
	lock := New(ctx, redisClient, key)
	defer lock.UnLock()

	err := lock.Lock()
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
}

// 测试加锁成功
func TestSevLockSuccess(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key_TestSevLockSuccess"

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock := New(ctx, redisClient, key)
		err := lock.Lock()
		if err != nil {
			t.Errorf("线程一：Lock() returned unexpected error: %v", err)
			return
		}
		time.Sleep(time.Second * 3)
		defer lock.UnLock()
		log.Println("线程一：执行结束")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		log.Println("线程二：开始抢夺锁资源")
		lock := New(ctx, redisClient, key)

		times, _ := redisClient.TTL(ctx, key).Result()
		log.Println("线程二：ttl 时间:", times.Milliseconds())

		err := lock.Lock()
		if err == nil {
			defer lock.UnLock()
			t.Errorf("线程二：Lock() returned unexpected error: %v", err)
			return
		}
	}()

	wg.Wait()
}

// 测试可重入锁计数器
func TestSevLockCounter(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key_TestSevLockCounter"
	lock := New(ctx, redisClient, key)

	err := lock.Lock()
	if err != nil {
		t.Errorf("任务1：Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock()

	err = lock.Lock()
	if err != nil {
		t.Errorf("任务2：Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock()
}

// 测试锁自动续期
func TestSevAutoRenewSuccess(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()

	key := "test_key_TestSevAutoRenewSuccess"
	token := "some_token"
	token2 := "some_token2"

	var wg sync.WaitGroup
	wg.Add(2)

	// 线程1
	go func() {
		defer wg.Done()
		lock := New(ctx, redisClient, key, WithToken(token), WithAutoRenew())
		err := lock.Lock()
		if err != nil {
			t.Errorf("Lock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock()

		log.Println("线程1：自旋锁加锁成功")
		time.Sleep(time.Second * 10)
		log.Println("线程1：任务执行结束")
	}()

	// 线程2
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 7)
		log.Println("线程2：开始抢夺锁资源")
		lock := New(ctx, redisClient, key, WithToken(token2), WithAutoRenew())
		err := lock.Lock()
		if err == nil {
			defer lock.UnLock()
			t.Errorf("线程2 Lock() not expectation success")
			return
		}
	}()

	wg.Wait()
}

func TestAutoRenew5(t *testing.T) {
	redisClient := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}
	lock := New(context.TODO(), redisClient, "key", WithToken("token"),
		WithAutoRenew())
	err := lock.Lock()
	require.NoError(t, err)
	defer lock.UnLock()

	time.Sleep(time.Second * 20)
}

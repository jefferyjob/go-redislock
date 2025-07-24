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
// docker run -d -p 63790:6379 --name go_redis_lock redis
// 注意：该服务在 GITHUB ACTIONS 并不会被测试
func getRedisClient() (RedisInter, *redis.Client) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return nil, nil
	}

	if true {
		return nil, nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:63790",
	})
	rdbAdapter := NewRedisV9Adapter(rdb)

	return rdbAdapter, rdb
}

func TestSevLock(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.TODO()
	key := "test_key_TestSevLock"
	lock := New(redisClient, key)
	defer lock.UnLock(ctx)

	err := lock.Lock(ctx)
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
}

// 测试加锁成功
func TestSevLockSuccess(t *testing.T) {
	redisClient, rdb := getRedisClient()
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
		lock := New(redisClient, key)
		err := lock.Lock(ctx)
		if err != nil {
			t.Errorf("线程一：Lock() returned unexpected error: %v", err)
			return
		}
		time.Sleep(time.Second * 3)
		defer lock.UnLock(ctx)
		log.Println("线程一：执行结束")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		log.Println("线程二：开始抢夺锁资源")
		lock := New(redisClient, key)

		times, _ := rdb.TTL(ctx, key).Result()
		log.Println("线程二：ttl 时间:", times.Milliseconds())

		err := lock.Lock(ctx)
		if err == nil {
			defer lock.UnLock(ctx)
			t.Errorf("线程二：Lock() returned unexpected error: %v", err)
			return
		}
	}()

	wg.Wait()
}

// 测试可重入锁计数器
func TestSevLockCounter(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key_TestSevLockCounter"
	lock := New(redisClient, key)

	err := lock.Lock(ctx)
	if err != nil {
		t.Errorf("任务1：Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock(ctx)

	err = lock.Lock(ctx)
	if err != nil {
		t.Errorf("任务2：Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock(ctx)
}

// 测试锁自动续期
func TestSevAutoRenewSuccess(t *testing.T) {
	redisClient, _ := getRedisClient()
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
		lock := New(redisClient, key, WithToken(token), WithAutoRenew())
		err := lock.Lock(ctx)
		if err != nil {
			t.Errorf("Lock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock(ctx)

		log.Println("线程1：自旋锁加锁成功")
		time.Sleep(time.Second * 10)
		log.Println("线程1：任务执行结束")
	}()

	// 线程2
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 7)
		log.Println("线程2：开始抢夺锁资源")
		lock := New(redisClient, key, WithToken(token2), WithAutoRenew())
		err := lock.Lock(ctx)
		if err == nil {
			defer lock.UnLock(ctx)
			t.Errorf("线程2 Lock() not expectation success")
			return
		}
	}()

	wg.Wait()
}

func TestAutoRenew5(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()

	lock := New(redisClient, "key", WithToken("token"),
		WithAutoRenew())
	err := lock.Lock(ctx)
	require.NoError(t, err)
	defer lock.UnLock(ctx)

	time.Sleep(time.Second * 20)
}

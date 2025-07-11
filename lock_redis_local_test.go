package go_redislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// 加锁成功，并执行业务代码
func TestLockSuccess(t *testing.T) {
	ctx := context.Background()
	key := "test_key_TestLockSuccess"
	token := "some_token"

	// 创建 redismock 客户端
	db, mock := redismock.NewClientMock()

	// 创建 RedisLock 实例
	lock := New(ctx, db, key, WithToken(token))

	// 设置模拟锁获取成功的行为
	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))

	err := lock.Lock()
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}

	defer lock.UnLock()

	// 在这里执行锁定期间的任务，确保任务可以在锁定期间正常执行
	// ...
	log.Println("Execute Business Code: start")
	fmt.Println("任务执行")
	log.Println("Execute Business Code: end")
}

// 加锁失败的情况
// 模拟思路如下：
// 线程1：同一个key加锁成功，执行任务5秒
// 线程2：使用此key继续加锁，期望加锁失败
func TestLockFail(t *testing.T) {
	key := "test_key_TestLockFail"
	token := "some_token"

	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))

	var wg sync.WaitGroup
	wg.Add(2)

	// 线程一：期望加锁成功
	go func() {
		defer wg.Done()
		log.Println("线程1: 启动")
		lock := New(ctx, db, key, WithToken(token))
		err := lock.Lock()
		if err == nil {
			log.Println("线程1: 加锁成功")
		} else {
			t.Errorf("线程1: Lock() returned unexpected error: %v", err)
			return
		}

		defer lock.UnLock()
		time.Sleep(time.Second * 5)
		log.Println("线程1: 任务执行")
	}()

	// 线程二：期望加锁失败
	go func() {
		defer wg.Done()
		log.Println("线程2: 启动")
		time.Sleep(time.Second * 2)
		lock := New(ctx, db, key, WithToken(token+"task-2"))
		err := lock.Lock()
		if err == nil {
			log.Println("线程2: 加锁成功")
			t.Errorf("线程2: Lock() Lock should have failed")
		} else {
			log.Println("线程2: 加锁失败")
			return
		}

		defer lock.UnLock()
		time.Sleep(time.Second * 5)
		log.Println("线程2: 任务执行")
	}()

	wg.Wait()
}

// 解锁失败的情况
// 逻辑一：使用 A Token 加锁
// 逻辑二：使用 B Token 为逻辑一加的锁解锁，期望：解锁失败
func TestUnlockFail(t *testing.T) {
	key := "test_key_TestUnlockFail"
	token := "some_token"
	ctx := context.Background()

	db, mock := redismock.NewClientMock()
	lock := New(ctx, db, key, WithToken(token))

	// 加锁逻辑
	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))

	err := lock.Lock()
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}

	// 解锁逻辑
	mock.ExpectEval(reentrantUnLockScript, []string{key}, token+"test-2").SetVal(int64(0)) // 模拟解锁失败

	err = lock.UnLock()

	if err == nil {
		t.Error("UnLock() expected to return an error, but got nil")
		return
	}
}

// 自旋锁成功的情况
// 线程一：自旋锁加锁成功，任务执行，最终解锁
// 线程二：自旋锁进行加锁，最终获得线程一交出的锁，执行任务
func TestSpinLockSuccess(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestSpinLockSuccess"
	token := "some_token"
	token2 := "some_token2"
	spinTimeout := time.Duration(5) * time.Second

	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	mock.ExpectEval(reentrantLockScript, []string{key}, token2, lockTime.Milliseconds()).SetVal(int64(1))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock := New(ctx, db, key, WithToken(token))
		err := lock.SpinLock(spinTimeout)
		if err != nil {
			t.Errorf("线程1：SpinLock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock()
		log.Println("线程1：自旋锁加锁成功")
		log.Println("线程1：任务执行结束")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		lock := New(ctx, db, key, WithToken(token2))
		err := lock.SpinLock(spinTimeout)
		if err != nil {
			t.Errorf("线程2：SpinLock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock()
		log.Println("线程2：自旋锁加锁成功")
		time.Sleep(time.Second * 2)
		log.Println("线程2：任务执行结束")
	}()

	wg.Wait()
}

// 自旋锁超时的情况
// 线程一：自旋锁加锁且执行任务 5 秒钟
// 线程二：自旋锁设置 3 秒的超时时间获取锁，期望：获取锁失败
func TestSpinLockTimeout(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestSpinLockTimeout"
	token := "some_token"
	token2 := "some_token2"
	spinTimeout := time.Duration(5) * time.Second
	spinTimeout2 := time.Duration(3) * time.Second

	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	mock.ExpectEval(reentrantLockScript, []string{key}, token2, lockTime.Milliseconds()).SetVal(int64(0))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock := New(ctx, db, key, WithToken(token))
		err := lock.SpinLock(spinTimeout)
		if err != nil {
			t.Errorf("线程1：SpinLock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock()
		log.Println("线程1：自旋锁加锁成功")
		time.Sleep(time.Second * 5)
		log.Println("线程1：任务执行结束")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		lock := New(ctx, db, key, WithToken(token2))
		err := lock.SpinLock(spinTimeout2)
		if err != nil {
			log.Println("线程2：获取锁成功")
			return
		}
		defer lock.UnLock()

		t.Errorf("线程2：SpinLock() Timeout error")
		return
	}()

	wg.Wait()
}

// 锁手动续期成功的情况
func TestRenewSuccess(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestRenewSuccess"
	token := "some_token"

	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	mock.ExpectEval(reentrantRenewScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))

	// 设置模拟锁续期成功的行为
	mock.ExpectExpire(key, lockTime).SetVal(true)

	lock := New(ctx, db, key, WithToken(token))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 3)
		err := lock.Renew()
		if err != nil {
			t.Errorf("Renew() returned unexpected error: %v", err)
			return
		}
		log.Println("锁续期成功")
	}()

	err := lock.Lock()
	if err != nil {
		t.Errorf("加锁失败: %v", err)
		return
	}
	defer lock.UnLock()

	// 模拟任务执行超过默认时间
	log.Println("任务执行：开始")
	time.Sleep(time.Second * 8)
	log.Println("任务执行：结束")

	wg.Wait()
}

// 锁手动续期失败的情况
func TestRenewFail(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestRenewFail"
	token := "some_token"

	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	// 设置模拟锁续期成功的行为
	mock.ExpectExpire(key, lockTime).SetVal(false)

	lock := New(ctx, db, key, WithToken(token))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 3)
		err := lock.Renew()
		if err == nil {
			t.Errorf("Renew() returned unexpected error: %v", err)
			return
		}
	}()

	err := lock.Lock()
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock()

	// 模拟任务执行超过默认时间
	log.Println("任务执行：开始")
	time.Sleep(time.Second * 8)
	log.Println("任务执行：结束")

	wg.Wait()
}

// 测试设置默认超时时间
// 线程1：设置默认超时时间为10
// 线程2：第7秒的时候抢夺抢夺线程1的锁，预期：抢夺失败
func TestWithTimeout(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestWithTimeout"
	token := "some_token"
	timeout := time.Duration(10) * time.Second

	mock.ExpectEval(reentrantLockScript, []string{key}, token, timeout.Milliseconds()).SetVal(int64(1))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock := New(ctx, db, key, WithToken(token), WithTimeout(timeout))
		err := lock.Lock()
		if err != nil {
			t.Errorf("线程1：Lock() returned unexpected error: %v", err)
			return
		}
		defer lock.UnLock()
		time.Sleep(time.Second * timeout)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 7)
		lock := New(ctx, db, key, WithToken(token), WithTimeout(timeout))
		err := lock.Lock()
		if err == nil {
			t.Errorf("线程2：Lock() expected lock failure")
			return
		}
	}()

	wg.Wait()
}

// 测试自动创建key和token
func TestLockCreateAutoKeyToken(t *testing.T) {
	ctx := context.Background()
	db, _ := redismock.NewClientMock()
	key := "test_key_TestLockCreateAutoKeyToken"
	New(ctx, db, key)
}

// 测试锁自动续期
func TestAutoRenew(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()

	key := "test_key_TestAutoRenew"
	token := "some_token"

	mock.ExpectEval(reentrantLockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	mock.ExpectEval(reentrantRenewScript, []string{key}, token, lockTime.Milliseconds()).SetVal(int64(1))
	mock.ExpectExpire(key, lockTime).SetVal(true)

	lock := New(ctx, db, key, WithToken(token), WithAutoRenew())
	err := lock.Lock()
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock()

	time.Sleep(time.Second * 6)
}

// 测试自动续期的 Context 操作取消
// func TestAutoRenewContextCancellation(t *testing.T) {
//	// 创建可取消的上下文
//	ctx, cancel := context.WithCancel(context.Background())
//
//	db, mock := redismock.NewClientMock()
//	key := "test_key"
//	token := "some_token"
//
//	// 设置模拟锁获取成功的行为
//	mock.ExpectEval(lockScript, []string{key}, token, lockTime.Milliseconds()).SetVal(1)
//	// 设置模拟锁续期成功的行为
//	mock.ExpectExpire(key, lockTime).SetVal(true)
//
//	// 创建 RedisLock 实例，开启自动续期
//	lock := New(ctx, db, key, WithToken(token), WithAutoRenew())
//
//	var wg sync.WaitGroup
//	wg.Add(1)
//
//	// 模拟锁自动续期的 goroutine
//	go func() {
//		defer wg.Done()
//		// 模拟锁续期成功
//		mock.ExpectEval(renewScript, []string{key}, token, lockTime.Milliseconds()).SetVal(1)
//		// 等待一段时间，模拟自动续期的过程
//		time.Sleep(time.Second * 3)
//		// 取消上下文，模拟上下文的取消
//		cancel()
//		// 等待 goroutine 退出
//		time.Sleep(time.Second * 2)
//	}()
//
//	err := lock.Lock()
//	if err != nil {
//		t.Errorf("Lock() 返回意外的错误: %v", err)
//		return
//	}
//	defer lock.UnLock()
//
//	// 等待一段时间，模拟锁自动续期的过程
//	time.Sleep(time.Second * 8)
//
//	wg.Wait()
// }

// Redis服务器测试
// 下面的代码将借助 redis 服务器进行测试，可以更加方便的测试服务中的问题
// 你可以实用下面的命令启动一个redis容器进行测试
// docker run -d -p 6379:6379 --name go_redis_lock redis
// 注意：该服务在 GITHUB ACTIONS 并不会被测试

func getRedisClient() *redis.Client {
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

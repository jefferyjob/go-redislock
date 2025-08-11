package adapter

import (
	"context"
	"fmt"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
	v9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
	"log"
	"sync"
	"testing"
	"time"
)

var (
	addr         = "127.0.0.1"
	port         = "63790"
	luaSetScript = `return redis.call("SET", KEYS[1], ARGV[1])`
	luaGetScript = `return redis.call("GET", KEYS[1])`
)

// Redis服务器测试
// 下面的代码将借助 redis 服务器进行测试，可以更加方便的测试服务中的问题
// 你可以实用下面的命令启动一个redis容器进行测试
// docker run -d -p 63790:6379 --name go_redis_lock redis
// 注意：该服务在 GITHUB ACTIONS 并不会被测试
func getRedisClient() (redislock.RedisInter, *v9.Client) {
	// if os.Getenv("GITHUB_ACTIONS") == "true" {
	// 	return nil, nil
	// }

	// if true {
	// 	return nil, nil
	// }

	rdb := v9.NewClient(&v9.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	})
	rdbAdapter := MustNew(rdb)

	return rdbAdapter, rdb
}

// 测试加锁流程
func TestSevLock(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	key := "test_key"

	ctx := context.TODO()
	lock := redislock.New(redisClient, key, redislock.WithAutoRenew())
	err := lock.Lock(ctx)
	if err != nil {
		t.Errorf("lock error: %v", err)
		return
	}
	defer lock.UnLock(ctx)

	// 模拟业务处理
	time.Sleep(5 * time.Second)
}

// 测试锁资源抢夺
// 线程1抢夺锁资源
// 线程2尝试抢夺锁资源失败
func TestSevLockSuccess(t *testing.T) {
	redisClient, rdb := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key"

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lock := redislock.New(redisClient, key)
		err := lock.Lock(ctx)
		if err != nil {
			t.Errorf("线程一：Lock() returned unexpected error: %v", err)
			return
		}
		time.Sleep(time.Second * 3)
		defer lock.UnLock(ctx)
		log.Println("线程一：任务执行结束")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 1)
		log.Println("线程二：开始抢夺锁资源")
		lock := redislock.New(redisClient, key)

		times, _ := rdb.TTL(ctx, "{"+key+"}").Result()
		log.Println("线程二：ttl 过期时间还有: ", times.Seconds(), " 秒")

		err := lock.Lock(ctx)
		if err == nil {
			defer lock.UnLock(ctx)
			t.Errorf("线程二：Lock() returned unexpected error: %v", err)
			return
		}

		log.Println("线程二：抢夺锁失败，锁已被其他线程占用")
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
	key := "test_key"
	lock := redislock.New(redisClient, key)

	err := lock.Lock(ctx)
	if err != nil {
		t.Errorf("任务1：error: %v", err)
		return
	}
	defer lock.UnLock(ctx)
	log.Println("任务1：锁已获取，开始执行任务")

	err = lock.Lock(ctx)
	if err != nil {
		t.Errorf("任务2： error: %v", err)
		return
	}
	defer lock.UnLock(ctx)
	log.Println("任务2：锁已获取，开始执行任务")
}

// 测试锁自动续期
// 测试线程1加锁，任务执行时间10秒，需要自动续期，保证不被其他资源抢占锁
// 线程2在第7秒尝试抢夺锁资源失败
func TestSevAutoRenewSuccess(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()

	key := "test_key"
	token := "test_token"
	token2 := "test_token2"

	var wg sync.WaitGroup
	wg.Add(2)

	// 线程1
	go func() {
		defer wg.Done()
		lock := redislock.New(redisClient, key, redislock.WithToken(token), redislock.WithAutoRenew())
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
		lock := redislock.New(redisClient, key, redislock.WithToken(token2), redislock.WithAutoRenew())
		err := lock.Lock(ctx)
		if err == nil {
			defer lock.UnLock(ctx)
			t.Errorf("线程2 Lock() not expectation success")
			return
		}
	}()

	wg.Wait()
}

// 展示自动续期流程
// 可以在协程里清晰的看到每 1/3 的时间自动续期锁
func TestSevAutoRenewList(t *testing.T) {
	redisClient, rdb := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	ctx := context.Background()
	key := "test_key"

	lock := redislock.New(redisClient, key, redislock.WithToken("token"),
		redislock.WithTimeout(10*time.Second),
		redislock.WithAutoRenew())
	err := lock.Lock(ctx)
	require.NoError(t, err)
	defer lock.UnLock(ctx)

	// 不断查询该锁的到期时间进行展示
	go func() {
		for i := 0; i < 20; i++ {
			times, _ := rdb.TTL(ctx, "{"+key+"}").Result()
			log.Println("ttl 过期时间还有: ", times.Seconds(), " 秒")
			time.Sleep(time.Second) // 每秒执行一次
		}
	}()

	// 模拟业务处理
	time.Sleep(time.Second * 20)
}

// ----------------------------------------------------------------------------------------------
// 测试适配器的兼容性
// ----------------------------------------------------------------------------------------------

// redis适配器测试
func TestSevNewRedisV9Adapter(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	adapter := MustNew(v9.NewClient(&v9.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	}))

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

// go-zero 适配器测试
func TestSevNewGoZeroAdapter(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	adapter := MustNew(zeroRdb.MustNewRedis(zeroRdb.RedisConf{
		Host: fmt.Sprintf("%s:%s", addr, port),
		Type: "node",
	}))

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

// gf v2 适配器测试
func TestSevNewGfV2Adapter(t *testing.T) {
	redisClient, _ := getRedisClient()
	if redisClient == nil {
		log.Println("Github actions skip this test")
		return
	}

	rdb, err := gfRdbV2.New(&gfRdbV2.Config{
		Address: fmt.Sprintf("%s:%s", addr, port),
	})
	require.NoError(t, err)

	adapter := MustNew(rdb)

	ctx := context.Background()
	key := "test_key"

	// 线程2抢占锁资源-预期失败
	go func() {
		time.Sleep(time.Second * 1)
		lock := redislock.New(adapter, key)
		err = lock.Lock(ctx)
		if err == nil {
			t.Errorf("Lock() returned unexpected success: %v", err)
			return
		}
		log.Println("线程2：抢占锁失败，锁已被其他线程占用")
	}()

	// 线程1加锁-预期成功
	lock := redislock.New(adapter, key)
	err = lock.Lock(ctx)
	if err != nil {
		t.Errorf("Lock() returned unexpected error: %v", err)
		return
	}
	defer lock.UnLock(ctx)

	// 模拟业务处理
	log.Println("线程1：锁已获取，开始执行任务")
	time.Sleep(time.Second * 5)
}

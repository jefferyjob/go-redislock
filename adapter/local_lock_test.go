// 连接redis服务 测试加锁和解锁
package adapter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
	adapterGfV2 "github.com/jefferyjob/go-redislock/adapter/gf/v2"
	"github.com/jefferyjob/go-redislock/adapter/gozero"
	v9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	gzRdb "github.com/zeromicro/go-zero/core/stores/redis"
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

	adapter := gozero.New(gzRdb.MustNewRedis(gzRdb.RedisConf{
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

	adapter := adapterGfV2.New(rdb)

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
	}
}

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
			inputKey:    "testKey",
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

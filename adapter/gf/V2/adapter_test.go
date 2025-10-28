package v2

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/database/gredis"
	redislock "github.com/jefferyjob/go-redislock"
)

var (
	addr         = "127.0.0.1"
	port         = "63790"
	luaSetScript = `return redis.call("SET", KEYS[1], ARGV[1])`
	luaGetScript = `return redis.call("GET", KEYS[1])`
	luaDelScript = `return redis.call("DEL", KEYS[1])`
)

func getRedisClient() (redislock.RedisInter, *gredis.Redis) {
	rdb, err := gredis.New(&gredis.Config{
		Address: fmt.Sprintf("%s:%s", addr, port),
	})
	if err != nil {
		return nil, nil
	}

	adapter := New(rdb)
	return adapter, rdb
}

// gf v2 适配器测试
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

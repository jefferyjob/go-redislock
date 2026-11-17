package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	redislock "github.com/jefferyjob/go-redislock"
	"github.com/redis/go-redis/v9"
)

// 锁安全性回归测试
//
// 固化安全修复的行为:
//  1. 主锁过期后,残留的重入计数器不允许凭空"重入"(否则互斥被破坏);
//  2. 锁丢失后,续期必须失败,不允许凭残留计数器误报续期成功;
//  3. 公平锁解锁非持有者的 requestId 必须失败,且不影响真正持有者;
//  4. 可重入锁部分解锁后锁仍被持有,完全解锁后锁消失。

// newRawClient 返回原生 go-redis 客户端,用于直接断言 Redis 中的锁状态。
// 注意:锁脚本会将 key 包上 hash tag,原生访问需使用 "{key}" 形式。
func newRawClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	})
}

// Test_OrphanCounter_NoPhantomReentry 主锁被其他客户端持有时,残留计数器不能误判为可重入。
// 修复前:本端凭孤儿计数器 INCR 并返回成功,导致两个客户端同时"持有"锁。
func Test_OrphanCounter_NoPhantomReentry(t *testing.T) {
	ctx := context.Background()
	rdb := newRawClient()
	defer rdb.Close()

	const rawKey = "reg_orphan_key"
	const key = "{reg_orphan_key}"
	const tokenA = "reg_token_a"
	const tokenB = "reg_token_b"
	cleanupKeys(t, rdb, key, key+":count:"+tokenA, key+":count:"+tokenB)

	// A 正常加锁,产生计数器 {key}:count:tokenA
	lockA := redislock.New(getRedisClient(), rawKey, redislock.WithToken(tokenA))
	if err := lockA.Lock(ctx); err != nil {
		t.Fatalf("A 加锁失败:%v", err)
	}

	// 模拟 TTL 边界:A 的主锁过期后被 B 抢走(计数器 key 的 TTL 略晚于主锁,可能残留)
	if err := rdb.Do(ctx, "SET", key, tokenB, "PX", 60000).Err(); err != nil {
		t.Fatalf("模拟 B 抢锁失败:%v", err)
	}

	// A 再次加锁:必须失败,不允许凭残留计数器"重入"
	if err := lockA.Lock(ctx); !errors.Is(err, redislock.ErrLockFailed) {
		t.Errorf("A 应加锁失败(孤儿计数器不可重入),实际 %v", err)
	}

	// A 解锁:必须失败,且不得删除 B 的锁
	if err := lockA.UnLock(ctx); !errors.Is(err, redislock.ErrUnLockFailed) {
		t.Errorf("A 应解锁失败,实际 %v", err)
	}
	got, err := rdb.Get(ctx, key).Result()
	if err != nil || got != tokenB {
		t.Errorf("B 的锁必须保持完好,期望 %q,实际 %q (err=%v)", tokenB, got, err)
	}

	// A 的孤儿计数器应被解锁脚本清理,避免后续误判
	if cnt, err := rdb.Get(ctx, key+":count:"+tokenA).Result(); !errors.Is(err, redis.Nil) {
		t.Errorf("A 的孤儿计数器应被清理,实际 %q (err=%v)", cnt, err)
	}

	cleanupKeys(t, rdb, key, key+":count:"+tokenA, key+":count:"+tokenB)
}

// Test_Renew_RejectsOrphanCounter 锁丢失后,续期必须失败,不得凭残留计数器误报成功。
// 修复前:续期脚本只看计数器,会在已失去锁的情况下返回成功(掩盖锁丢失),
// 且自动续期 goroutine 会永久续活孤儿计数器、永远不退出。
func Test_Renew_RejectsOrphanCounter(t *testing.T) {
	ctx := context.Background()
	rdb := newRawClient()
	defer rdb.Close()

	const rawKey = "reg_renew_key"
	const key = "{reg_renew_key}"
	const token = "reg_token"
	cleanupKeys(t, rdb, key, key+":count:"+token)

	lock := redislock.New(getRedisClient(), rawKey, redislock.WithToken(token))
	if err := lock.Lock(ctx); err != nil {
		t.Fatalf("加锁失败:%v", err)
	}

	// 模拟锁过期丢失(如持有者冻结超过 TTL),计数器残留
	if err := rdb.Del(ctx, key).Err(); err != nil {
		t.Fatalf("模拟锁丢失失败:%v", err)
	}

	// 续期:必须失败
	if err := lock.Renew(ctx); !errors.Is(err, redislock.ErrLockRenewFailed) {
		t.Errorf("锁丢失后续期应失败,实际 %v", err)
	}

	// 续期失败不得续活孤儿计数器(计数器应保持自然过期,不被 PEXPIRE 刷新)
	if ttl, err := rdb.PTTL(ctx, key+":count:"+token).Result(); err == nil && ttl < 0 {
		// TTL < 0 表示 key 不存在(已过期)或无过期时间(-1);
		// 残留计数器带 TTL,正常应为 0 < ttl <= 初值,这里只防被续期脚本重置为 -1
		if ttl == -1 {
			t.Errorf("孤儿计数器不应被续期脚本改为永不过期")
		}
	}

	cleanupKeys(t, rdb, key, key+":count:"+token)
}

// Test_FairUnLock_NotHolder 公平锁解锁非持有者的 requestId 必须失败,且不影响真正持有者。
// 修复前:解锁脚本无条件返回成功,调用方会误以为释放成功。
func Test_FairUnLock_NotHolder(t *testing.T) {
	ctx := context.Background()
	rdb := newRawClient()
	defer rdb.Close()

	const rawKey = "reg_fair_key"
	const key = "{reg_fair_key}"
	const reqA = "reg_req_a"
	const reqB = "reg_req_b"
	cleanupKeys(t, rdb, key, key+":queue")

	lockA := redislock.New(getRedisClient(), rawKey)
	if err := lockA.FairLock(ctx, reqA); err != nil {
		t.Fatalf("A 公平加锁失败:%v", err)
	}

	// B(非持有者)解锁:必须失败,且不得删除 A 的锁
	lockB := redislock.New(getRedisClient(), rawKey)
	if err := lockB.FairUnLock(ctx, reqB); !errors.Is(err, redislock.ErrUnLockFailed) {
		t.Errorf("非持有者解锁应失败,实际 %v", err)
	}
	got, err := rdb.Get(ctx, key).Result()
	if err != nil || got != reqA {
		t.Errorf("A 的锁必须保持完好,期望 %q,实际 %q (err=%v)", reqA, got, err)
	}

	// A 解锁:必须成功
	if err := lockA.FairUnLock(ctx, reqA); err != nil {
		t.Errorf("持有者解锁应成功,实际 %v", err)
	}

	cleanupKeys(t, rdb, key, key+":queue")
}

// Test_ReentrantPartialUnlock 可重入锁部分解锁后锁仍被持有,完全解锁后锁消失。
func Test_ReentrantPartialUnlock(t *testing.T) {
	ctx := context.Background()
	rdb := newRawClient()
	defer rdb.Close()

	const rawKey = "reg_reentrant_key"
	const key = "{reg_reentrant_key}"
	const token = "reg_token"
	cleanupKeys(t, rdb, key, key+":count:"+token)

	lock := redislock.New(getRedisClient(), rawKey, redislock.WithToken(token))
	if err := lock.Lock(ctx); err != nil {
		t.Fatalf("首次加锁失败:%v", err)
	}
	if err := lock.Lock(ctx); err != nil {
		t.Fatalf("重入加锁失败:%v", err)
	}

	// 第一次解锁:部分释放,锁必须仍存在
	if err := lock.UnLock(ctx); err != nil {
		t.Fatalf("部分解锁失败:%v", err)
	}
	got, err := rdb.Get(ctx, key).Result()
	if err != nil || got != token {
		t.Errorf("部分解锁后锁必须仍被持有,期望 %q,实际 %q (err=%v)", token, got, err)
	}

	// 第二次解锁:完全释放
	if err := lock.UnLock(ctx); err != nil {
		t.Fatalf("完全解锁失败:%v", err)
	}
	if _, err := rdb.Get(ctx, key).Result(); !errors.Is(err, redis.Nil) {
		t.Errorf("完全解锁后锁应消失,实际 %q (err=%v)", got, err)
	}

	cleanupKeys(t, rdb, key, key+":count:"+token)
}

// cleanupKeys 清理测试键,避免残留影响其他测试
func cleanupKeys(t *testing.T, rdb *redis.Client, keys ...string) {
	t.Helper()
	if err := rdb.Del(context.Background(), keys...).Err(); err != nil {
		t.Fatalf("清理测试键失败:%v", err)
	}
}

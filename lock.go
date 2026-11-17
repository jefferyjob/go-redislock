package go_redislock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RedisInter Redis 客户端接口
type RedisInter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd
}

// RedisCmd Eval 返回结果的接口
type RedisCmd interface {
	Result() (interface{}, error)
	Int64() (int64, error)
}

// type RedisInter interface {
// 	redis.Scripter
// }

// RedisLockInter defines the interface for distributed Redis locks
type RedisLockInter interface {
	// Lock 加锁
	Lock(ctx context.Context) error
	// SpinLock 自旋锁。
	SpinLock(ctx context.Context, timeout time.Duration) error
	// UnLock 解锁
	UnLock(ctx context.Context) error
	// Renew 锁续期
	Renew(ctx context.Context) error

	// FairLock 公平锁加锁
	FairLock(ctx context.Context, requestId string) error
	// SpinFairLock 自旋公平锁
	SpinFairLock(ctx context.Context, requestId string, timeout time.Duration) error
	// FairUnLock 公平锁解锁
	FairUnLock(ctx context.Context, requestId string) error
	// FairRenew 公平锁续期
	FairRenew(ctx context.Context, requestId string) error

	// RLock 读锁加锁
	RLock(ctx context.Context) error
	// RUnLock 读锁解锁
	RUnLock(ctx context.Context) error
	// SpinRLock 自旋读锁
	SpinRLock(ctx context.Context, timeout time.Duration) error
	// RRenew 读锁续期
	RRenew(ctx context.Context) error

	// WLock 写锁加锁
	WLock(ctx context.Context) error
	// WUnLock 写锁解锁
	WUnLock(ctx context.Context) error
	// SpinWLock 自旋写锁
	SpinWLock(ctx context.Context, timeout time.Duration) error
	// WRenew 写锁续期
	WRenew(ctx context.Context) error
}

type RedisLock struct {
	redis          RedisInter
	key            string
	token          string
	lockTimeout    time.Duration
	isAutoRenew    bool
	requestTimeout time.Duration

	// mu 保护 autoRenew 的并发读写（加锁与解锁可能来自不同 goroutine）
	// mu protects concurrent access to autoRenew (lock and unlock may come from different goroutines)
	mu        sync.Mutex
	autoRenew *renewHandle
}

// renewHandle 标识一次自动续期 goroutine 的启动，用于安全地比较与清理
// renewHandle identifies one auto-renew goroutine start, for safe comparison and cleanup
type renewHandle struct {
	cancel context.CancelFunc
}

type Option func(lock *RedisLock)

// New creates a RedisLock instance
func New(redisClient RedisInter, lockKey string, options ...Option) RedisLockInter {
	lock := &RedisLock{
		redis:          redisClient,
		lockTimeout:    lockTime,       // 锁默认超时时间
		requestTimeout: requestTimeout, // 公平锁在队列中的最大等待时间
	}

	for _, f := range options {
		f(lock)
	}
	lock.key = lockKey

	// 非正数的超时配置回退到默认值，避免无效的 Redis TTL 或续期 ticker panic
	// Non-positive timeout values fall back to defaults to avoid invalid Redis TTL or renew ticker panic
	if lock.lockTimeout <= 0 {
		lock.lockTimeout = lockTime
	}
	if lock.requestTimeout <= 0 {
		lock.requestTimeout = requestTimeout
	}

	// 如果未设置锁的Token，则生成一个唯一的Token
	if lock.token == "" {
		lock.token = fmt.Sprintf("lock_token:%s", uuid.New().String())
	}

	return lock
}

// WithTimeout sets the expiration time of the lock.
// Non-positive values are ignored and the default is used.
//
// WithTimeout 设置锁的过期时间。非正值会被忽略并回退到默认值
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew enables automatic lock renewal.
// The renewal goroutine renews every lockTimeout/3 until the lock is fully released
// (via UnLock/FairUnLock/RUnLock/WUnLock) or a renewal fails (e.g. the lock was lost).
// It runs on a context detached from the one passed to Lock: cancelling that context
// does NOT stop renewal, so a caller must unlock to release the lock.
//
// WithAutoRenew 是否开启自动续期。
// 续期 goroutine 每 lockTimeout/3 续期一次，直到锁被完全释放
// （通过 UnLock/FairUnLock/RUnLock/WUnLock）或续期失败（如锁已丢失）为止。
// 续期运行在独立于加锁时传入 ctx 的 context 上：取消该 ctx 不会停止续期，
// 因此调用方必须通过解锁来释放锁。
func WithAutoRenew() Option {
	return func(lock *RedisLock) {
		lock.isAutoRenew = true
	}
}

// WithToken sets a custom token for the lock instance.
// The token identifies the lock holder and MUST be unique among all processes/instances
// competing for the same lock key. Reusing the same token across processes shares the
// reentrant counter and allows multiple holders to acquire the lock simultaneously,
// breaking mutual exclusion. Leave it empty to use a random UUID.
//
// WithToken 设置锁的 Token，用于标识当前持有者。
// Token 在竞争同一锁 key 的所有进程/实例间必须唯一。多个进程复用同一 Token
// 会共享可重入计数器，导致多方同时持锁、破坏互斥。留空则使用随机 UUID。
func WithToken(token string) Option {
	return func(lock *RedisLock) {
		lock.token = token
	}
}

// WithRequestTimeout sets the maximum wait time in the fair lock queue.
// Non-positive values are ignored and the default is used.
//
// WithRequestTimeout 设置公平锁在队列中的最大等待时间。非正值会被忽略并回退到默认值
func WithRequestTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.requestTimeout = timeout
	}
}

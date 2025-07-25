package go_redislock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// RedisLockInter defines the interface for distributed Redis locks
// RedisLockInter 定义了 Redis 分布式锁的接口
type RedisLockInter interface {

	// Lock tries to acquire a standard lock.
	// This implementation supports "reentrant locks". If the lock is currently held by the same key+token, reentry is allowed and the count is increased. Unlock() needs to be called a corresponding number of times to release the lock.
	//
	// Lock 尝试获取普通锁。
	// 该实现支持“可重入锁”，如果当前已由相同 key+token 持有，允许重入并增加计数。需调用相应次数 Unlock() 释放
	Lock(ctx context.Context) error

	// SpinLock keeps trying to acquire the lock until timeout.
	// SpinLock 在指定超时时间内不断尝试加锁。
	SpinLock(ctx context.Context, timeout time.Duration) error

	// UnLock releases the standard lock.
	// If it is a reentrant lock, each call will reduce the holding count until the count reaches 0 and the lock will be released.
	//
	// UnLock 释放普通锁。
	// 如果为重入锁，每调用一次减少一次持有计数，直到计数为 0 锁会被释放
	UnLock(ctx context.Context) error

	// Renew manually extends the lock expiration.
	// Renew 手动延长锁的有效期。
	Renew(ctx context.Context) error

	// FairLock tries to acquire a fair lock using the given requestId.
	// 公平锁加锁：使用指定的 requestId 获取公平锁。
	FairLock(ctx context.Context, requestId string) error

	// SpinFairLock keeps trying to acquire a fair lock until timeout.
	// SpinFairLock 在指定超时时间内不断尝试获取公平锁。
	SpinFairLock(ctx context.Context, requestId string, timeout time.Duration) error

	// FairUnLock releases the fair lock held by the given requestId.
	// FairUnLock 根据 requestId 释放公平锁。
	FairUnLock(ctx context.Context, requestId string) error

	// FairRenew manually extends the expiration of a fair lock.
	// FairRenew 手动延长指定 requestId 的公平锁有效期。
	FairRenew(ctx context.Context, requestId string) error
}

type RedisLock struct {
	redis           RedisInter
	key             string
	token           string
	lockTimeout     time.Duration
	isAutoRenew     bool
	requestTimeout  time.Duration
	autoRenewCancel context.CancelFunc
}

type Option func(lock *RedisLock)

// New creates a RedisLock instance
// If a token is not provided via WithToken, a unique token will be automatically generated and an implementation of RedisLockInter will be returned
//
// Parameters:
// - ctx: context for locking operations and cancellations
// - redisClient: abstract Redis client that implements RedisInter
// - lockKey: Redis key used for locking
// - options: optional configuration items, such as timeout, automatic renewal, etc.
//
// New 创建一个 RedisLock 实例
// 如果未通过 WithToken 提供令牌，则将自动生成一个唯一的令牌，最终返回 RedisLockInter 的一个实现
//
// 参数：
// - rdb：实现 RedisInter 的抽象 Redis 客户端
// - lockKey：用于锁定的 Redis 键
// - options：可选配置项，如超时时间、自动续期等
func New(rdb RedisInter, lockKey string, options ...Option) RedisLockInter {
	lock := &RedisLock{
		redis:          rdb,
		lockTimeout:    lockTime,       // 锁默认超时时间
		requestTimeout: requestTimeout, // 公平锁在队列中的最大等待时间
	}

	for _, f := range options {
		f(lock)
	}
	lock.key = lockKey

	// 如果未设置锁的Token，则生成一个唯一的Token
	if lock.token == "" {
		lock.token = fmt.Sprintf("lock_token:%s", uuid.New().String())
	}

	return lock
}

// WithTimeout sets the expiration time of the lock
// WithTimeout 设置锁的过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew enables automatic lock renewal
// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *RedisLock) {
		lock.isAutoRenew = true
	}
}

// WithToken sets a custom token for the lock instance
// WithToken 设置锁的 Token，用于标识当前持有者
func WithToken(token string) Option {
	return func(lock *RedisLock) {
		lock.token = token
	}
}

// WithRequestTimeout sets the maximum wait time in the fair lock queue
// WithRequestTimeout 设置公平锁在队列中的最大等待时间
func WithRequestTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.requestTimeout = timeout
	}
}

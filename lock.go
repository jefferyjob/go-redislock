package go_redislock

import (
	"context"
	"fmt"
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

	// MultiLock 联锁加锁
	// MultiLock(ctx context.Context, locks []RedisLockInter) error
	// MultiUnLock 联锁解锁
	// MultiUnLock(ctx context.Context, locks []RedisLockInter) error
	// SpinMultiLock 自旋联锁
	// SpinMultiLock(ctx context.Context, locks []RedisLockInter, timeout time.Duration) error
	// MultiRenew 联锁续期
	// MultiRenew(ctx context.Context, locks []RedisLockInter) error
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

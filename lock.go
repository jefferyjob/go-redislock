package go_redislock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// 默认锁超时时间
const lockTime = 5 * time.Second
const lockQueueTimeout = 60 * time.Second // 超时机制设置为 60 秒

type RedisLockInter interface {
	// Lock 可重入锁
	Lock() error
	SpinLock(timeout time.Duration) error // 自旋锁
	UnLock() error
	Renew() error

	// FairLock 公平锁
	FairLock() error
	FairUnLock() error
	FairRenew() error

	// MultiLock 联锁
	MultiLock(lockKeys []string) error
	MultiUnLock(lockKeys []string) error
	MultiRenew(lockKeys []string) error

	// RedLock 红锁
	RedLock(lockKeys []string) error
	RedUnLock(lockKeys []string) error
	RedRenew(lockKeys []string) error
}

type RedisInter interface {
	redis.Scripter
}

type RedisLock struct {
	context.Context
	redis           RedisInter
	key             string
	token           string
	lockTimeout     time.Duration
	isAutoRenew     bool
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
	mutex           sync.Mutex
}

type Option func(lock *RedisLock)

func New(ctx context.Context, redisClient RedisInter, lockKey string, options ...Option) RedisLockInter {
	lock := &RedisLock{
		Context:     ctx,
		redis:       redisClient,
		lockTimeout: lockTime,
	}
	for _, f := range options {
		f(lock)
	}

	lock.key = lockKey

	// token 自动生成
	if lock.token == "" {
		lock.token = fmt.Sprintf("lock_token:%s", uuid.New().String())
	}

	return lock
}

// WithTimeout 设置锁过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *RedisLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *RedisLock) {
		lock.isAutoRenew = true
	}
}

// WithToken 设置锁的Token
func WithToken(token string) Option {
	return func(lock *RedisLock) {
		lock.token = token
	}
}

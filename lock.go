package go_redislock

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisLockInter interface {
	// Lock 加锁
	Lock() error
	// SpinLock 自旋锁
	SpinLock(timeout time.Duration) error
	// UnLock 解锁
	UnLock() error
	// Renew 手动续期
	Renew() error
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
	// automatically generate tokens
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

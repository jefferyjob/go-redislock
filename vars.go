package go_redislock

import (
	"errors"
	"time"
)

const (
	// 默认锁超时时间
	lockTime = 5 * time.Second
)

var (
	// ErrLockFailed 加锁失败
	ErrLockFailed = errors.New("lock failed")
	// ErrUnLockFailed 解锁失败
	ErrUnLockFailed = errors.New("unLock failed")
	// ErrSpinLockTimeout 自旋锁加锁超时
	ErrSpinLockTimeout = errors.New("spin lock timeout")
	// ErrSpinLockDone 自旋锁加锁超时
	ErrSpinLockDone = errors.New("spin lock context done")
	// ErrLockRenewFailed 锁续期失败
	ErrLockRenewFailed = errors.New("lock renew failed")
	// ErrException 内部异常
	ErrException = errors.New("go redis lock internal exception")
)

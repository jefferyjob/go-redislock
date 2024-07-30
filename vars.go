package go_redislock

import "errors"

var (
	// ErrScriptFailed Lua脚本执行失败
	ErrScriptFailed = errors.New("script execution failed")

	// ErrSpinLockDone 自旋锁加锁超时
	ErrSpinLockDone = errors.New("spin lock context done")

	// ErrLockFailed 加锁失败
	ErrLockFailed = errors.New("lock failed")
	// ErrUnLockFailed 解锁失败
	ErrUnLockFailed = errors.New("unLock failed")
	// ErrSpinLockTimeout 自旋锁加锁超时
	ErrSpinLockTimeout = errors.New("spin lock timeout")
	// ErrException 内部异常
	ErrException = errors.New("go-redislock internal exception")
)

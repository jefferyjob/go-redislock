package go_redislock

import "errors"

var (
	// ErrScriptException Lua脚本执行异常
	ErrScriptException = errors.New("script execution exception")
	// ErrScriptFailed Lua脚本执行失败
	ErrScriptFailed = errors.New("script execution failed")
	// ErrSpinLockTimeout 自旋锁加锁超时
	ErrSpinLockTimeout = errors.New("spin lock timeout")
	// ErrSpinLockDone 自旋锁加锁超时
	ErrSpinLockDone = errors.New("spin lock context done")
)

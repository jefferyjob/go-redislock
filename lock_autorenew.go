package go_redislock

import (
	"context"
	"log"
	"time"
)

// startAutoRenew starts the auto-renew goroutine (at most one per lock instance).
// On reentrant acquisition it reuses the running goroutine instead of overwriting
// the cancel func, which would leak the previous renewal goroutine.
// The goroutine runs on a context detached from the caller's: its lifecycle is
// controlled by stopAutoRenew (unlock) or self-exit on renewal failure.
//
// startAutoRenew 启动自动续期 goroutine（每个实例同一时间至多一个）。
// 可重入加锁时若已有续期 goroutine 在运行则直接复用，避免覆盖取消函数导致旧 goroutine 泄漏。
// goroutine 运行在独立于调用方的 context 上：生命周期由 stopAutoRenew（解锁）或续期失败（自退出）控制。
func (l *RedisLock) startAutoRenew(renew func(ctx context.Context) error) {
	if !l.isAutoRenew {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.autoRenew != nil {
		// 已有续期 goroutine 在运行（可重入场景），复用即可
		// A renewal goroutine is already running (reentrant case), reuse it
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	handle := &renewHandle{cancel: cancel}
	l.autoRenew = handle
	go func() {
		// goroutine 自行退出（如续期失败）时清理句柄，允许下次加锁重新启动
		// Clear the handle on self-exit (e.g. renewal failed) so a later lock can restart
		defer l.clearAutoRenew(handle)
		runRenewLoop(ctx, l.lockTimeout/3, renew)
	}()
}

// stopAutoRenew stops the auto-renew goroutine.
//
// stopAutoRenew 停止自动续期 goroutine。
func (l *RedisLock) stopAutoRenew() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.autoRenew != nil {
		l.autoRenew.cancel()
		l.autoRenew = nil
	}
}

// clearAutoRenew clears the handle only if it still belongs to the given
// goroutine, so a self-exiting goroutine cannot cancel a newer one.
//
// clearAutoRenew 清理自动续期句柄，仅当仍属于本次启动的 goroutine 时生效，
// 避免自行退出的旧 goroutine 误清新启动的续期。
func (l *RedisLock) clearAutoRenew(handle *renewHandle) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.autoRenew == handle {
		l.autoRenew = nil
	}
}

// runRenewLoop renews the lock at a fixed interval; it logs and exits on the
// first renewal failure (lock lost or internal error).
//
// runRenewLoop 按固定周期执行续期，续期失败（锁丢失或内部异常）时记录日志并退出。
func runRenewLoop(ctx context.Context, interval time.Duration, renew func(ctx context.Context) error) {
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := renew(ctx); err != nil {
				log.Printf("go-redislock: auto renew failed: %v", err)
				return
			}
		}
	}
}

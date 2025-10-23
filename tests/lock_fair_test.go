package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
)

func Test_FairLock(t *testing.T) {
	adapter := getRedisClient()

	tests := []struct {
		name       string
		inputKey   string
		inputReqId string
		sleepTime  time.Duration // 模拟业务执行时间
		wantErr    error
	}{
		{
			name:       "公平锁-加锁成功",
			inputKey:   "test_key",
			inputReqId: "test_req",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := redislock.New(adapter, tt.inputKey,
				redislock.WithAutoRenew(),
				redislock.WithRequestTimeout(lockTime), // 设置公平锁请求超时时间
			)

			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
			defer lock.FairUnLock(ctx, tt.inputReqId)

			if tt.sleepTime != time.Duration(0) {
				// 模拟业务执行时间
				time.Sleep(tt.sleepTime)
			}
		})
	}
}

package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	redislock "github.com/jefferyjob/go-redislock"
	"github.com/stretchr/testify/require"
)

func Test_FairLock(t *testing.T) {
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

			lock := redislock.New(getRedisClient(), tt.inputKey,
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

func Test_FairUnLock(t *testing.T) {
	tests := []struct {
		name          string
		inputKey      string
		inputReqId    string
		wantLockErr   error
		wantUnlockErr error
	}{
		{
			name:          "公平锁-解锁成功",
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			wantLockErr:   nil,
			wantUnlockErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := redislock.New(getRedisClient(), tt.inputKey, redislock.WithAutoRenew())
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantLockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantLockErr)
			}

			err = lock.FairUnLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantUnlockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantUnlockErr)
			}
		})
	}
}

func Test_FairRenew(t *testing.T) {
	tests := []struct {
		name         string
		inputKey     string
		inputReqId   string
		sleepTime    time.Duration // 模拟业务执行时间
		wantErr      error
		wantRenewErr error
	}{
		{
			name:         "公平锁-续期成功",
			inputKey:     "key",
			inputReqId:   "req_id",
			sleepTime:    2 * time.Second,
			wantErr:      nil,
			wantRenewErr: nil,
		},
		{
			name:         "公平锁-续期失败",
			inputKey:     "key",
			inputReqId:   "req_id",
			sleepTime:    6 * time.Second,
			wantErr:      nil,
			wantRenewErr: redislock.ErrLockRenewFailed, // 续期失败
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := redislock.New(getRedisClient(), tt.inputKey)
			err := lock.FairLock(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
			defer lock.FairUnLock(ctx, tt.inputReqId)

			// 模拟业务执行时间
			if tt.sleepTime != time.Duration(0) {
				time.Sleep(tt.sleepTime)
			}

			err = lock.FairRenew(ctx, tt.inputReqId)
			if !errors.Is(err, tt.wantRenewErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_SpinFairLock(t *testing.T) {
	tests := []struct {
		name        string
		inputKey    string
		inputReqId  string
		spinTimeout time.Duration
		before      func()
		wantErr     error
	}{
		{
			name:        "公平锁-自旋锁-加锁成功",
			inputKey:    "test_key",
			inputReqId:  "test_req",
			spinTimeout: 3 * time.Second,
			wantErr:     nil,
		},
		{
			name:        "公平锁-自旋锁-加锁失败",
			inputKey:    "test_key",
			inputReqId:  "test_req",
			spinTimeout: 3 * time.Second,
			before: func() {
				ctx := context.Background()

				go func() {
					lock := redislock.New(getRedisClient(), "test_key_1")
					err := lock.SpinFairLock(ctx, "test_req_1", 5*time.Second)
					require.NoError(t, err)
					defer lock.FairUnLock(ctx, "test_req_1")
					time.Sleep(1 * time.Second) // 模拟业务执行时间
				}()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			lock := redislock.New(getRedisClient(), tt.inputKey)
			err := lock.SpinFairLock(ctx, tt.inputReqId, tt.spinTimeout)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
			defer lock.FairUnLock(ctx, tt.inputReqId)
		})
	}
}

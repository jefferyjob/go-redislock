package go_redislock

import (
	"context"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

// 公平锁错误测试
func TestFairLockErr(t *testing.T) {
	tests := []struct {
		name         string
		mock         func(t *testing.T) *redis.Client
		inputKey     string
		inputReqId   string
		inputTimeout float64
		wantErr      error
	}{
		{
			name: "Lua脚本执行异常",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).
					SetErr(ErrException)
				return db
			},
			inputKey:     "test_key",
			inputReqId:   "test_req_id",
			inputTimeout: 5 * time.Second.Seconds(),
			wantErr:      ErrException,
		},
		{
			name: "公平锁-获取锁资源失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).
					SetErr(ErrLockFailed)
				return db
			},
			inputKey:     "test_key",
			inputReqId:   "test_req_id",
			inputTimeout: 5 * time.Second.Seconds(),
			wantErr:      ErrLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := New(context.TODO(), tt.mock(t), tt.inputKey)
			err := lock.FairLock(tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package go_redislock

import (
	"context"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestFairLock(t *testing.T) {
	tests := []struct {
		name         string
		mock         func(t *testing.T) *redis.Client
		inputKey     string
		inputReqId   string
		inputTimeout float64
		wantErr      error
	}{
		{
			name: "Lua脚本执行成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).
					SetVal(int64(1))
				return db
			},
			inputKey:     "test_key",
			inputReqId:   "test_req_id",
			inputTimeout: 5 * time.Second.Seconds(),
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := New(context.TODO(), tt.mock(t), tt.inputKey, WithAutoRenew())
			err := lock.FairLock(tt.inputReqId)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFairUnLock(t *testing.T) {
	tests := []struct {
		name          string
		mock          func(t *testing.T) *redis.Client
		inputKey      string
		inputReqId    string
		inputTimeout  float64
		wantLockErr   error
		wantUnlockErr error
	}{
		{
			name: "加锁成功后解锁成功",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").SetVal(int64(1))
				return db
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			inputTimeout:  5 * time.Second.Seconds(),
			wantLockErr:   nil,
			wantUnlockErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := New(context.TODO(), tt.mock(t), tt.inputKey, WithAutoRenew())
			err := lock.FairLock(tt.inputReqId)
			if !errors.Is(err, tt.wantLockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantLockErr)
			}

			err = lock.FairUnLock(tt.inputReqId)
			if !errors.Is(err, tt.wantUnlockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantUnlockErr)
			}
		})
	}
}

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

// 公平锁解锁错误测试
func TestFairUnLockErr(t *testing.T) {
	tests := []struct {
		name          string
		mock          func(t *testing.T) *redis.Client
		inputKey      string
		inputReqId    string
		inputTimeout  float64
		wantLockErr   error
		wantUnlockErr error
	}{
		{
			name: "公平锁-解锁-Lua脚本执行异常",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").SetErr(ErrException)
				return db
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			inputTimeout:  5 * time.Second.Seconds(),
			wantLockErr:   nil,
			wantUnlockErr: ErrException,
		},
		{
			name: "公平锁-解锁-解锁失败",
			mock: func(t *testing.T) *redis.Client {
				db, mock := redismock.NewClientMock()
				mock.ExpectEval(fairLockScript, []string{"test_key"},
					"test_req_id", 5*time.Second.Seconds(), 5*time.Second.Seconds()).SetVal(int64(1))
				mock.ExpectEval(fairUnLockScript, []string{"test_key"}, "test_req_id").SetErr(ErrUnLockFailed)
				return db
			},
			inputKey:      "test_key",
			inputReqId:    "test_req_id",
			inputTimeout:  5 * time.Second.Seconds(),
			wantLockErr:   nil,
			wantUnlockErr: ErrUnLockFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := New(context.TODO(), tt.mock(t), tt.inputKey)
			err := lock.FairLock(tt.inputReqId)
			if !errors.Is(err, tt.wantLockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantLockErr)
			}

			err = lock.FairUnLock(tt.inputReqId)
			if !errors.Is(err, tt.wantUnlockErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantUnlockErr)
			}
		})
	}
}

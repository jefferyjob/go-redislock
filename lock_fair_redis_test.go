package go_redislock

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestFairLock(t *testing.T) {
	tests := []struct {
		name      string
		mock      func(t *testing.T) *redis.Client
		requestId string
		wantErr   error
	}{
		{
			name: "Lua脚本执行异常",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lock := New(context.TODO(), tt.mock(t), "fair_lock_test")
			err := lock.FairLock(tt.requestId)
			if errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

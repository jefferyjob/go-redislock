package go_redislock

import (
	"context"
	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	mockV7 "github.com/go-redis/redismock/v7"
	mockV8 "github.com/go-redis/redismock/v8"
	mockV9 "github.com/go-redis/redismock/v9"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	v9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
	"testing"
)

// ----------------------------------------------------------------------------------------------
//  Redis Mock 适配器 Start
// ----------------------------------------------------------------------------------------------

type RedisMockAdapter struct {
	client *v9.Client
}

func NewRedisMockAdapter(client *v9.Client) RedisInter {
	return &RedisMockAdapter{client: client}
}

func (r *RedisMockAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd := r.client.Eval(ctx, script, keys, args...)
	return &RedisMockCmdWrapper{cmd: cmd}
}

type RedisMockCmdWrapper struct {
	cmd *v9.Cmd
}

func (w *RedisMockCmdWrapper) Result() (interface{}, error) {
	return w.cmd.Result()
}

func (w *RedisMockCmdWrapper) Int64() (int64, error) {
	return w.cmd.Int64()
}

// ----------------------------------------------------------------------------------------------
//  Redis Mock 适配器 End
// ----------------------------------------------------------------------------------------------

func TestNewRedisAdapter(t *testing.T) {
	v7Client := &v7.Client{}
	v8Client := &v8.Client{}
	v9Client := &v9.Client{}
	zeroClient := &zeroRdb.Redis{}
	gfClient := &gfRdbV2.Redis{}

	tests := []struct {
		name    string
		client  interface{}
		wantErr bool
	}{
		{"redis.v7", v7Client, false},
		{"redis.v8", v8Client, false},
		{"redis.v9", v9Client, false},
		{"go-zero redis", zeroClient, false},
		{"gf redis v2", gfClient, false},
		{"unsupported", "xxx", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewRedisAdapter(tt.client)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
			}
		})
	}
}

func TestRedisV9Adapter(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		wantInt64  int64
		wantErr    error
	}{
		{
			name: "Redis Adapter",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := mockV9.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				return NewRedisV9Adapter(db)
			},
			inputKey:   "key",
			inputToken: "token",
			wantInt64:  1,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rdbAdapter := tt.mock(t, ctx)

			cmd := rdbAdapter.Eval(ctx, reentrantLockScript, []string{tt.inputKey}, tt.inputToken, lockTime.Milliseconds())
			_, err := cmd.Result()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			i, err := cmd.Int64()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			if i != tt.wantInt64 {
				t.Errorf("Expected %d, got %d", tt.wantInt64, i)
			}
		})
	}
}

func TestRedisV8Adapter(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		wantInt64  int64
		wantErr    error
	}{
		{
			name: "Redis Adapter",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := mockV8.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				return NewRedisV8Adapter(db)
			},
			inputKey:   "key",
			inputToken: "token",
			wantInt64:  1,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rdbAdapter := tt.mock(t, ctx)

			cmd := rdbAdapter.Eval(ctx, reentrantLockScript, []string{tt.inputKey}, tt.inputToken, lockTime.Milliseconds())
			_, err := cmd.Result()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			i, err := cmd.Int64()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			if i != tt.wantInt64 {
				t.Errorf("Expected %d, got %d", tt.wantInt64, i)
			}
		})
	}
}

func TestRedisV7Adapter(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) RedisInter
		inputKey   string
		inputToken string
		wantInt64  int64
		wantErr    error
	}{
		{
			name: "Redis Adapter",
			mock: func(t *testing.T, ctx context.Context) RedisInter {
				db, mock := mockV7.NewClientMock()
				mock.ExpectEval(reentrantLockScript, []string{"key"}, "token", lockTime.Milliseconds()).
					SetVal(int64(1))
				return NewRedisV7Adapter(db)
			},
			inputKey:   "key",
			inputToken: "token",
			wantInt64:  1,
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rdbAdapter := tt.mock(t, ctx)

			cmd := rdbAdapter.Eval(ctx, reentrantLockScript, []string{tt.inputKey}, tt.inputToken, lockTime.Milliseconds())
			_, err := cmd.Result()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			i, err := cmd.Int64()
			if tt.wantErr != nil {
				require.NoError(t, err)
			}

			if i != tt.wantInt64 {
				t.Errorf("Expected %d, got %d", tt.wantInt64, i)
			}
		})
	}
}

func TestGoZeroRedisAdapter(t *testing.T) {
}

func TestGoFrameRedisV2Adapter(t *testing.T) {

}

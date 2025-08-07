package go_redislock

import (
	"context"
	mockV7 "github.com/go-redis/redismock/v7"
	mockV8 "github.com/go-redis/redismock/v8"
	mockV9 "github.com/go-redis/redismock/v9"
	"github.com/jefferyjob/go-redislock/adapter"
	v9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
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
				return adapter.MustNew(db)
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
				return adapter.MustNew(db)
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
				return adapter.MustNew(db)
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

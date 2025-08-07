package v8

import (
	"context"
	mockV8 "github.com/go-redis/redismock/v8"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	evalLua = "test lua script"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		mock       func(t *testing.T, ctx context.Context) redislock.RedisInter
		inputKey   string
		inputToken string
		wantInt64  int64
		wantErr    error
	}{
		{
			name: "Redis Adapter",
			mock: func(t *testing.T, ctx context.Context) redislock.RedisInter {
				db, mock := mockV8.NewClientMock()
				mock.ExpectEval(evalLua, []string{"key"}, "token", 5*time.Millisecond).
					SetVal(int64(1))
				return New(db)
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

			cmd := rdbAdapter.Eval(ctx, evalLua, []string{tt.inputKey}, tt.inputToken, 5*time.Millisecond)
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

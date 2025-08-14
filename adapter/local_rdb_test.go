// 连接redis服务测试 Eval 命令
package adapter

import (
	"context"
	"fmt"
	gfRdbV1 "github.com/gogf/gf/database/gredis"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	adapterGfV1 "github.com/jefferyjob/go-redislock/adapter/gf/v1"
	adapterGfV2 "github.com/jefferyjob/go-redislock/adapter/gf/v2"
	"github.com/jefferyjob/go-redislock/adapter/gozero"
	"github.com/stretchr/testify/require"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
	"strconv"
	"testing"
)

func TestSevNewGoZero(t *testing.T) {
	adapter := gozero.New(zeroRdb.MustNewRedis(zeroRdb.RedisConf{
		Host: fmt.Sprintf("%s:%s", addr, port),
		Type: "node",
	}))

	ctx := context.Background()

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	r, err := cmd.Result()
	require.NoError(t, err)
	val, ok := r.(int64)
	if !ok {
		t.Errorf("result is not int64, got type %T", r)
	}
	// 判断是否等于 1
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

func TestSevNewGfV1(t *testing.T) {
	newPort, _ := strconv.Atoi(port)
	rdb := gfRdbV1.New(&gfRdbV1.Config{
		Host: addr,
		Port: newPort,
	})

	ctx := context.Background()

	adapter := adapterGfV1.New(rdb)
	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	r, err := cmd.Result()
	require.NoError(t, err)
	val, ok := r.(int64)
	if !ok {
		t.Errorf("result is not int64, got type %T", r)
	}
	// 判断是否等于 1
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

func TestSevNewGfV2(t *testing.T) {
	rdb, err := gfRdbV2.New(&gfRdbV2.Config{
		Address: fmt.Sprintf("%s:%s", addr, port),
	})
	require.NoError(t, err)

	ctx := context.Background()
	adapter := adapterGfV2.New(rdb)

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	r, err := cmd.Result()
	require.NoError(t, err)
	val, ok := r.(int64)
	if !ok {
		t.Errorf("result is not int64, got type %T", r)
	}
	// 判断是否等于 1
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

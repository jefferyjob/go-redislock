// 连接redis服务测试 Eval 命令
package adapter

import (
	"context"
	"fmt"
	rdbV7 "github.com/go-redis/redis/v7"
	rdbV8 "github.com/go-redis/redis/v8"
	gfRdbV1 "github.com/gogf/gf/database/gredis"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	adapterGfV1 "github.com/jefferyjob/go-redislock/adapter/gf/v1"
	adapterGfV2 "github.com/jefferyjob/go-redislock/adapter/gf/v2"
	"github.com/jefferyjob/go-redislock/adapter/gozero"
	adapterRdbV7 "github.com/jefferyjob/go-redislock/adapter/v7"
	adapterRdbV8 "github.com/jefferyjob/go-redislock/adapter/v8"
	adapterRdbV9 "github.com/jefferyjob/go-redislock/adapter/v9"
	rdbV9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
	"strconv"
	"testing"
)

func TestSevNewRdbV7(t *testing.T) {
	adapter := adapterRdbV7.New(rdbV7.NewClient(&rdbV7.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	}))

	ctx := context.Background()

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	_, err := cmd.Result()
	require.NoError(t, err)

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

func TestSevNewRdbV8(t *testing.T) {
	adapter := adapterRdbV8.New(rdbV8.NewClient(&rdbV8.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	}))

	ctx := context.Background()

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	_, err := cmd.Result()
	require.NoError(t, err)

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

func TestSevNewRdbV9(t *testing.T) {
	adapter := adapterRdbV9.New(rdbV9.NewClient(&rdbV9.Options{
		Addr: fmt.Sprintf("%s:%s", addr, port),
	}))

	ctx := context.Background()

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	_, err := cmd.Result()
	require.NoError(t, err)

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

func TestSevNewGoZero(t *testing.T) {
	adapter := gozero.New(zeroRdb.MustNewRedis(zeroRdb.RedisConf{
		Host: fmt.Sprintf("%s:%s", addr, port),
		Type: "node",
	}))

	ctx := context.Background()

	adapter.Eval(ctx, luaSetScript, []string{"test_key"}, "1")
	cmd := adapter.Eval(ctx, luaGetScript, []string{"test_key"})

	_, err := cmd.Result()
	require.NoError(t, err)

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

	_, err := cmd.Result()

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

	_, err = cmd.Result()
	require.NoError(t, err)

	i, err := cmd.Int64()
	require.NoError(t, err)
	if i != 1 {
		t.Errorf("Expected value 1, got %d", i)
		return
	}
}

package go_zero

import (
	"context"
	"fmt"

	redislock "github.com/jefferyjob/go-redislock"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RdbAdapter struct {
	client *redis.Redis
}

func New(client *redis.Redis) redislock.RedisInter {
	return &RdbAdapter{client: client}
}

// Eval 通过 go-zero 的 EvalCtx 执行 Lua 脚本，结果由 GoZeroRdbCmdWrapper 封装
func (r *RdbAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) redislock.RedisCmd {
	cmd, err := r.client.EvalCtx(ctx, script, keys, args...)
	return &RdbCmdWrapper{
		cmd: cmd,
		err: err,
	}
}

type RdbCmdWrapper struct {
	cmd interface{}
	err error
}

func (w *RdbCmdWrapper) Result() (interface{}, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.cmd, nil
}
func (w *RdbCmdWrapper) Int64() (int64, error) {
	if w.err != nil {
		return 0, w.err
	}

	switch v := w.cmd.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(v, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert result to int: %T", w.cmd)
	}
}

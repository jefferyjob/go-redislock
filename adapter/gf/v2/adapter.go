package v2

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	redislock "github.com/jefferyjob/go-redislock"
)

type RdbAdapter struct {
	client *gredis.Redis
}

func New(client *gredis.Redis) redislock.RedisInter {
	return &RdbAdapter{client: client}
}

func (r *RdbAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) redislock.RedisCmd {
	eval, err := r.client.Eval(ctx, script, int64(len(keys)), keys, args)
	return &RdbCmdWrapper{
		cmd: eval,
		err: err,
	}
}

type RdbCmdWrapper struct {
	cmd *g.Var
	err error
}

func (w *RdbCmdWrapper) Result() (interface{}, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.cmd.Val(), nil
}

func (w *RdbCmdWrapper) Int64() (int64, error) {
	if w.err != nil {
		return 0, w.err
	}

	switch v := w.cmd.Val().(type) {
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

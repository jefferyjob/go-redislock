package gf

import (
	"context"

	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	redislock "github.com/jefferyjob/go-redislock"
)

type RdbAdapter struct {
	client *gredis.Redis
}

func New(client *gredis.Redis) redislock.RedisInter {
	return &RdbAdapter{client: client}
}

func (r *RdbAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) redislock.RedisCmd {
	params := make([]interface{}, 0, 2+len(keys)+len(args))
	params = append(params, script, int64(len(keys)))
	for _, k := range keys {
		params = append(params, k)
	}
	params = append(params, args...)

	eval, err := r.client.DoVar("EVAL", params...)
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

	return w.cmd.Int64(), nil
}

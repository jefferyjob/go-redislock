package go_redislock

import (
	"context"
	v9 "github.com/redis/go-redis/v9"
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

package go_redislock

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// ----------------------------------------------------------------------------------------------
//  Redis Mock 适配器 Start
// ----------------------------------------------------------------------------------------------

type RedisMockAdapter struct {
	client *redis.Client
}

func NewRedisMockAdapter(client *redis.Client) RedisInter {
	return &RedisMockAdapter{client: client}
}

func (r *RedisMockAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd := r.client.Eval(ctx, script, keys, args...)
	return &RedisMockCmdWrapper{cmd: cmd}
}

type RedisMockCmdWrapper struct {
	cmd *redis.Cmd
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

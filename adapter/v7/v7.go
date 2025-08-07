package v9

import (
	"context"
	"github.com/go-redis/redis/v7"
	redislock "github.com/jefferyjob/go-redislock"
)

type RedisAdapter struct {
	client *redis.Client
}

func New(client *redis.Client) redislock.RedisInter {
	return &RedisAdapter{client: client}
}

func (r *RedisAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) redislock.RedisCmd {
	cmd := r.client.Eval(script, keys, args...)
	return &RedisCmdWrapper{cmd: cmd}
}

type RedisCmdWrapper struct {
	cmd *redis.Cmd
}

func (w *RedisCmdWrapper) Result() (interface{}, error) {
	return w.cmd.Result()
}
func (w *RedisCmdWrapper) Int64() (int64, error) {
	return w.cmd.Int64()
}

package go_redislock

import (
	"context"
	"fmt"

	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	v9 "github.com/redis/go-redis/v9"

	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
)

func NewRedisAdapter(rawClient interface{}) (RedisInter, error) {
	switch client := rawClient.(type) {
	case *v7.Client:
		return NewRedisV7Adapter(client), nil
	case *v8.Client:
		return NewRedisV8Adapter(client), nil
	case *v9.Client:
		return NewRedisV9Adapter(client), nil
	case *zeroRdb.Redis:
		return NewGoZeroRdbAdapter(client), nil
	default:
		return nil, fmt.Errorf("unsupported redis client type: %T", rawClient)
	}
}

// ---------- Redis v9 适配器 ----------

type RedisV9Adapter struct {
	client *v9.Client
}

func NewRedisV9Adapter(client *v9.Client) RedisInter {
	return &RedisV9Adapter{client: client}
}

func (r *RedisV9Adapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd := r.client.Eval(ctx, script, keys, args...)
	return &RedisV9CmdWrapper{cmd: cmd}
}

type RedisV9CmdWrapper struct {
	cmd *v9.Cmd
}

func (w *RedisV9CmdWrapper) Result() (interface{}, error) {
	return w.cmd.Result()
}
func (w *RedisV9CmdWrapper) Int64() (int64, error) {
	return w.cmd.Int64()
}

// ---------- Redis v8 适配器 ----------

type RedisV8Adapter struct {
	client *v8.Client
}

func NewRedisV8Adapter(client *v8.Client) RedisInter {
	return &RedisV8Adapter{client: client}
}

func (r *RedisV8Adapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd := r.client.Eval(ctx, script, keys, args...)
	return &RedisV8CmdWrapper{cmd: cmd}
}

type RedisV8CmdWrapper struct {
	cmd *v8.Cmd
}

func (w *RedisV8CmdWrapper) Result() (interface{}, error) {
	return w.cmd.Result()
}
func (w *RedisV8CmdWrapper) Int64() (int64, error) {
	return w.cmd.Int64()
}

// ---------- Redis v7 适配器 ----------

type RedisV7Adapter struct {
	client *v7.Client
}

func NewRedisV7Adapter(client *v7.Client) RedisInter {
	return &RedisV7Adapter{client: client}
}

func (r *RedisV7Adapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd := r.client.Eval(script, keys, args...)
	return &RedisV7CmdWrapper{cmd: cmd}
}

type RedisV7CmdWrapper struct {
	cmd *v7.Cmd
}

func (w *RedisV7CmdWrapper) Result() (interface{}, error) {
	return w.cmd.Result()
}
func (w *RedisV7CmdWrapper) Int64() (int64, error) {
	return w.cmd.Int64()
}

// ---------- go-zero Redis 适配器 ----------

type GoZeroRdbAdapter struct {
	client *zeroRdb.Redis
}

func NewGoZeroRdbAdapter(client *zeroRdb.Redis) RedisInter {
	return &GoZeroRdbAdapter{client: client}
}

// Eval 通过 go-zero 的 EvalCtx 执行 Lua 脚本，结果由 GoZeroRdbCmdWrapper 封装
func (r *GoZeroRdbAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	cmd, err := r.client.EvalCtx(ctx, script, keys, args...)
	return &GoZeroRdbCmdWrapper{
		cmd: cmd,
		err: err,
	}
}

type GoZeroRdbCmdWrapper struct {
	cmd interface{}
	err error
}

func (w *GoZeroRdbCmdWrapper) Result() (interface{}, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.cmd, nil
}
func (w *GoZeroRdbCmdWrapper) Int64() (int64, error) {
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

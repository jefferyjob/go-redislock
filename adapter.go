package go_redislock

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	v9 "github.com/redis/go-redis/v9"
)

// RedisInter Redis 客户端接口
type RedisInter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd
}

// RedisCmd Eval 返回结果的接口
type RedisCmd interface {
	Result() (interface{}, error)
	Int64() (int64, error)
}

// NewRedisAdapter creates a corresponding Redis adapter based on the passed in Redis client instance
//
// Parameters:
// - rawClient: Redis client instance
//
// Return value:
// - RedisInter interface instance, encapsulating a unified Eval method
// - error: If the passed in client type is not supported, an error is returned
//
// # NewRedisAdapter 根据传入的 Redis 客户端实例，创建对应的 Redis 适配器
//
// 参数：
//   - rawClient：Redis 客户端实例
//
// 返回值：
//   - RedisInter 接口实例，封装了统一的 Eval 方法
//   - error：若传入的客户端类型不受支持，则返回错误
func NewRedisAdapter(rawClient interface{}) (RedisInter, error) {
	switch client := rawClient.(type) {
	case *v7.Client:
		return NewRedisV7Adapter(client), nil
	case *v8.Client:
		return NewRedisV8Adapter(client), nil
	case *v9.Client:
		return NewRedisV9Adapter(client), nil
	default:
		return nil, fmt.Errorf("unsupported redis client type: %T", rawClient)
	}
}

func MustNewRedisAdapter(rawClient interface{}) RedisInter {
	adapter, err := NewRedisAdapter(rawClient)
	if err != nil {
		msg := fmt.Sprintf("%+v\n\n%s", err.Error(), debug.Stack())
		log.Println(msg)
		panic(msg)
	}
	return adapter
}

// ----------------------------------------------------------------------------------------------
// Redis v9 Adapter
// ----------------------------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------------------------
// Redis v8 Adapter
// ----------------------------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------------------------
// Redis v7 Adapter
// ----------------------------------------------------------------------------------------------

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

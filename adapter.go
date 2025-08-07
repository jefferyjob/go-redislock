package go_redislock

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	v9 "github.com/redis/go-redis/v9"

	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"

	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	gV2 "github.com/gogf/gf/v2/frame/g"
)

// RedisInter Redis 客户端接口
// type RedisInter interface {
// 	Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd
// }
//
// // RedisCmd Eval 返回结果的接口
// type RedisCmd interface {
// 	Result() (interface{}, error)
// 	Int64() (int64, error)
// }

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
	case *zeroRdb.Redis:
		return NewGoZeroRdbAdapter(client), nil
	case *gfRdbV2.Redis:
		return NewGfRedisV2Adapter(client), nil
	default:
		return nil, fmt.Errorf("unsupported redis client type: %T", rawClient)
	}
}

// MustNewRedisAdapter creates a Redis adapter and panics if an error occurs.
//
// It is a helper function that wraps NewRedisAdapter, and should be used when
// the caller expects the Redis client to be valid and initialization should not fail.
//
// Parameters:
// - rawClient: Redis client instance
//
// Return value:
// - RedisInter: The initialized Redis adapter instance
//
// Panic:
//   - If the provided client type is unsupported, the function logs the error and panics,
//     including the error message and current stack trace.
//
// # MustNewRedisAdapter 创建 Redis 适配器，如出错则直接 panic
//
// 它是对 NewRedisAdapter 的封装，适用于调用方确信 Redis 客户端类型合法且初始化不应失败的场景。
//
// 参数：
//   - rawClient：Redis 客户端实例
//
// 返回值：
//   - RedisInter：已初始化的 Redis 适配器实例
//
// 异常：
//   - 若传入的客户端类型不受支持，函数会打印错误信息和当前堆栈信息，并触发 panic
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

// ----------------------------------------------------------------------------------------------
// go-zero Redis Adapter
// ----------------------------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------------------------
// GoFrame Redis (gredis) v2 Adapter
// ----------------------------------------------------------------------------------------------

type GfRedisV2Adapter struct {
	client *gfRdbV2.Redis
}

func NewGfRedisV2Adapter(client *gfRdbV2.Redis) RedisInter {
	return &GfRedisV2Adapter{client: client}
}

func (r *GfRedisV2Adapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd {
	eval, err := r.client.Eval(ctx, script, int64(len(keys)), keys, args)
	return &GfRedisV2CmdWrapper{
		cmd: eval,
		err: err,
	}
}

type GfRedisV2CmdWrapper struct {
	cmd *gV2.Var
	err error
}

func (w *GfRedisV2CmdWrapper) Result() (interface{}, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.cmd.Val(), nil
}

func (w *GfRedisV2CmdWrapper) Int64() (int64, error) {
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

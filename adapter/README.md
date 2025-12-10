# Redis 客户端适配器
`go-redislock` 提供了可扩展的客户端适配机制，并内置支持多个主流 Redis 客户端。 如需完整示例，请参考 [examples](../examples)。

如果您当前使用的 Redis 客户端未被支持，也可以通过实现 `RedisInter` 接口来自定义适配器。

## 📦 导入适配器
请选择与您当前使用的 Redis 客户端版本相匹配的适配器：

```bash
# go-redis v9
go get -u github.com/jefferyjob/go-redislock/adapter/go-redis/V9

# go-redis v8
go get -u github.com/jefferyjob/go-redislock/adapter/go-redis/V8

# go-redis v7
go get -u github.com/jefferyjob/go-redislock/adapter/go-redis/V7

# go-zero
go get -u github.com/jefferyjob/go-redislock/adapter/go-zero/V1
```

## ❓ 没有适配器符合你的客户端
如果内置适配器无法满足需求，只需实现以下接口即可接入任何 Redis 客户端：

```go
// RedisInter 定义 Redis 客户端的最小能力集
type RedisInter interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) RedisCmd
}

// RedisCmd 定义 Eval 返回结果的通用访问形式
type RedisCmd interface {
	Result() (interface{}, error)
	Int64() (int64, error)
}
```

实现以上接口后即可直接与 `go-redislock` 联动。

## 🛠 示例：自定义 Goframe gredis 适配器
以下示例展示如何将 Goframe 的 `gredis` 客户端封装为可用于 `go-redislock` 的 Redis 适配器：

```go
package adapter

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
```

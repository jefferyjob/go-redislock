# go-redislock

[![Go](https://img.shields.io/badge/Go->=1.24-green)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/jefferyjob/go-redislock.svg)](https://github.com/jefferyjob/go-redislock/releases)
[![Action](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml/badge.svg)](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml)
[![Report](https://goreportcard.com/badge/github.com/jefferyjob/go-redislock)](https://goreportcard.com/report/github.com/jefferyjob/go-redislock)
[![Coverage](https://codecov.io/gh/jefferyjob/go-redislock/branch/main/graph/badge.svg)](https://codecov.io/gh/jefferyjob/go-redislock)
[![Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jefferyjob/go-redislock)
[![License](https://img.shields.io/github/license/jefferyjob/go-redislock)](https://github.com/jefferyjob/go-redislock/blob/main/LICENSE)

English | [简体中文](README.cn.md)

## Introduce
go-redislock is a library for Go that provides distributed lock functionality using Redis as the backend storage. Ensure data sharing and resource mutual exclusion under concurrent access in a distributed environment. Our distributed lock has the characteristics of reliability, high performance, timeout mechanism, reentrancy and flexible lock release method, which simplifies the use of distributed lock and allows you to focus on the realization of business logic.

## Redis client support
go-redislock provides a highly scalable client adaptation mechanism, and has built-in support for the following mainstream Redis clients. For detailed examples, please refer to [examples](examples/adapter) .

| Redis Client Version | Package path | Supported | Adapter method |
|----------------------|--------------------------------------------------| -------- |-----------------------|
| go-redis v7          | `github.com/go-redis/redis/v7`                   | ✅        | NewRedisV7Adapter()   |
| go-redis v8          | `github.com/go-redis/redis/v8`                   | ✅        | NewRedisV8Adapter()   |
| go-redis v9          | `github.com/redis/go-redis/v9`                   | ✅        | NewRedisV9Adapter()   |
| go-zero Redis        | `github.com/zeromicro/go-zero/core/stores/redis` | ✅        | NewGoZeroRdbAdapter() |
| goframe Redis        | `github.com/gogf/gf/v2/database/gredis`          | ✅        | NewGfRedisV2Adapter() |

If the Redis client you are using is not in the above list, you can also implement the interface `RedisInter` to connect to any Redis client.

## Quick start

### Install
```bash
go get -u github.com/jefferyjob/go-redislock
```

### Use Demo
```go
package main

import (
	"context"
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Create a Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create a Redis client adapter
	// Note: Use different adapters according to different redis client packages
	rdbAdapter := redislock.NewRedisV9Adapter(rdb)

	// Create a context for canceling lock operations
	ctx := context.Background()

	// Create a RedisLock object
	lock := redislock.New(rdbAdapter, "test_key")

	// acquire lock
	err := lock.Lock(ctx)
	if err != nil {
		fmt.Println("lock acquisition failed：", err)
		return
	}
	defer lock.UnLock(ctx) // unlock

	// Perform tasks during lockdown
	// ...
	fmt.Println("task execution completed")
}

```

### Configuration options
| **Option function** | **Description** | **Default value** |
| ----------------------------------- |------------------|---------|
| WithTimeout(d time.Duration) | Lock timeout (TTL) | 5s |
| WithAutoRenew() | Whether to automatically renew | false |
| WithToken(token string) | Reentrant lock Token (unique identifier) | Random UUID |
| WithRequestTimeout(d time.Duration) | Maximum waiting time for fair lock queue | Same as TTL |

### API Quick Check
```go
type RedisLockInter interface {
    // Lock Locking
    Lock(ctx context.Context) error
    // SpinLock Spinlock
    SpinLock(ctx context.Context, timeout time.Duration) error
    // UnLock Unlocking
    UnLock(ctx context.Context) error
    // Renew Manual renewal
    Renew(ctx context.Context) error
    
    // FairLock Fair lock locking
    FairLock(ctx context.Context, requestId string) error
    // SpinFairLock Spin Fair Lock
    SpinFairLock(ctx context.Context, requestId string, timeout time.Duration) error
    // FairUnLock Fair Lock Unlock
    FairUnLock(ctx context.Context, requestId string) error
    // FairRenew Fair Lock Renew
    FairRenew(ctx context.Context, requestId string) error
}
```


## Precautions
- Create a new `RedisLock` instance each time you lock.
- Please make sure your Redis server is set up correctly, connected and running properly.
- When using the automatic renewal function, ensure that there is no long-term blocking during task execution, so as not to cause the renewal to fail.
- Consider using appropriate timeout settings to avoid deadlocks due to network issues etc.
- Try to ensure that the same key is used to acquire and release locks to ensure correctness.
- In the process of using locks, it is recommended to carefully design and test critical code blocks to avoid race conditions and deadlock problems.

## License
This library is licensed under the MIT. See the LICENSE file for details.


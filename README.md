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

We implemented the following key capabilities:

- 🔒 Standard distributed locks (reentrant)
- 🔁 Spin locks
- ⚖️ Fair locks (FIFO order)
- 🔄 Manual and automatic renewal
- ✅ Compatibility with multiple Redis clients (v7/v8/v9, go-zero, goframe)

---

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

## Core Function Overview

| Function Type | Method Name | Description |
| ----- | --------------------------------------- |-----------------|
| Basic Lock Function | `Lock(ctx)` | Acquire a normal lock (supports reentrancy) |
| | `SpinLock(ctx, timeout)` | Acquire a normal lock by spin |
| | `UnLock(ctx)` | Unlock |
| | `Renew(ctx)` | Manually renew |
| Fair Lock Function | `FairLock(ctx, requestId)` | Acquire a fair lock (based on a FIFO queue) |
| | `SpinFairLock(ctx, requestId, timeout)` | Acquire a fair lock by spin |
| | `FairUnLock(ctx, requestId)` | Unlock a fair lock |
| | `FairRenew(ctx, requestId)` | Renew a fair lock |

The corresponding interface definitions are as follows

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


## Precautions
- It is recommended to use a new lock instance each time you acquire a lock.
- The same key and token must be used for locking and unlocking.
- The default TTL is 5 seconds, but it is recommended to set it based on the duration of the task.
- Automatic renewal is suitable for non-blocking tasks to avoid long blocking times.
- Fair locks require a unique request ID (UUID is recommended).
- Read locks can be concurrent, while write locks are mutually exclusive to avoid read-write conflicts.
- Failure of any sublock in an interlock releases all locked resources.
- Redis must remain available to avoid deadlocks caused by network issues.
- It is recommended to use defer unlocking in critical logic to prevent leaks.
- It is recommended to log or monitor lock acquisition failures, retries, and other behaviors.

## License
This library is licensed under the MIT. See the LICENSE file for details.


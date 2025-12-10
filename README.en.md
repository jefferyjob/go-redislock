# go-redislock

[![Go](https://img.shields.io/badge/Go->=1.21-green)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/jefferyjob/go-redislock.svg)](https://github.com/jefferyjob/go-redislock/releases)
[![Action](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml/badge.svg)](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml)
[![Report](https://goreportcard.com/badge/github.com/jefferyjob/go-redislock)](https://goreportcard.com/report/github.com/jefferyjob/go-redislock)
[![Coverage](https://codecov.io/gh/jefferyjob/go-redislock/branch/main/graph/badge.svg)](https://codecov.io/gh/jefferyjob/go-redislock)
[![Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jefferyjob/go-redislock)
[![License](https://img.shields.io/github/license/jefferyjob/go-redislock)](https://github.com/jefferyjob/go-redislock/blob/main/LICENSE)

English | [简体中文](README.md)

## Introduce
go-redislock is a library for Go that provides distributed lock functionality using Redis as the backend storage. Ensure data sharing and resource mutual exclusion under concurrent access in a distributed environment. Our distributed lock has the characteristics of reliability, high performance, timeout mechanism, reentrancy and flexible lock release method, which simplifies the use of distributed lock and allows you to focus on the realization of business logic.

We implemented the following key capabilities:

- 🔒 Standard distributed locks (reentrant)
- 🔁 Spin locks
- ⚖️ Fair locks (FIFO order)
- 🧵Read lock (multiple readers access concurrently, mutually exclusive writers)
- ✍️Write lock (exclusive access to a resource)
- 🔄 Manual and automatic renewal
- ✅ Compatibility with multiple Redis clients (v7/v8/v9, go-zero, goframe)

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
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v9"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Create a Redis client adapter
	rdbAdapter := adapter.New(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}))

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
### Normal Lock
| Method Name | Description |
|------------------------------|------------------------|
| `Lock(ctx)` | Acquire a normal lock (supports reentrancy) |
| `SpinLock(ctx, timeout)` | Acquire a normal lock using a spinlock method |
| `UnLock(ctx)` | Unlock operation |
| `Renew(ctx)` | Manual renewal |

### Fair Lock (FIFO)
| Method Name | Description |
|--------------------------------------------|----------------------|
| `FairLock(ctx, requestId)` | Acquire a fair lock (FIFO) |
| `SpinFairLock(ctx, requestId, timeout)` | Acquire a fair lock using a spinlock method |
| `FairUnLock(ctx, requestId)` | Unlock a fair lock |
| `FairRenew(ctx, requestId)` | Fair Lock Renewal |

### Read Lock
| Method Name | Description |
|--------------------------|-------------|
| `RLock(ctx)` | Acquire a read lock (supports reentrancy) |
| `SpinRLock(ctx, timeout)` | Acquire a read lock using a spinlock |
| `UnLRock(ctx)` | Unlock operation |
| `RRenew(ctx)` | Manually renew the lock |

### Write Lock
| Method Name | Description |
|--------------------------|-------------|
| `WLock(ctx)` | Acquire a write lock (supports reentrancy) |
| `SpinWLock(ctx, timeout)` | Acquire a write lock using a spinlock |
| `UnWLock(ctx)` | Unlock operation |
| `WRenew(ctx)` | Manually renew the lock |

### The interface is defined as follows
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

    // RLock read lock locked
    RLock(ctx context.Context) error
    // RUnLock read lock unlocked
    RUnLock(ctx context.Context) error
    // SpinRLock spin read lock
    SpinRLock(ctx context.Context, timeout time.Duration) error
    // RRenew read lock renewed
    RRenew(ctx context.Context) error
    
    // WLock write lock locked
    WLock(ctx context.Context) error
    // WUnLock write lock unlocked
    WUnLock(ctx context.Context) error
    // SpinWLock spin write lock
    SpinWLock(ctx context.Context, timeout time.Duration) error
    // WRenew write lock renewed
    WRenew(ctx context.Context) error
}
```

## Redis client support
go-redislock provides a highly scalable client adaptation mechanism, and has built-in support for the following mainstream Redis clients. For detailed examples, please refer to [examples](examples/adapter) .

| Redis Client Version | Package path                                             | Supported |
|------------------|----------------------------------------------------------| -------- |
| go-redis v7      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V7` | ✅        |
| go-redis v8      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V8` | ✅        | 
| go-redis v9      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V9` | ✅        | 
| go-zero redis    | `github.com/jefferyjob/go-redislock/adapter/go-zero/V1`  | ✅        | 
| goframe v1 redis | `github.com/jefferyjob/go-redislock/adapter/gf/V1`       | ✅        |
| goframe v2 redis | `github.com/jefferyjob/go-redislock/adapter/gf/V2`       | ✅        |

If the Redis client you are using is not in the above list, you can also implement the interface `RedisInter` to connect to any Redis client.


## Precautions
- It is recommended to use a new lock instance each time you acquire a lock.
- The same key and token must be used for locking and unlocking.
- The default TTL is 5 seconds, and it is recommended to set it based on the duration of the task.
- Automatic renewal is suitable for non-blocking tasks to avoid long blocking times.
- It is recommended to use `defer unlock` in critical logic to prevent leaks.
- It is recommended to log or monitor lock acquisition failures, retries, and other behaviors.
- Fair locks require a unique requestId (UUID is recommended).
- Read locks can be concurrent, while write locks are mutually exclusive to avoid read-write conflicts.
- If any sub-lock in the interlock fails, the successfully acquired lock will be released.
- There is a risk of deadlock if Redis is unavailable.

## License
This library is licensed under the MIT. See the LICENSE file for details.


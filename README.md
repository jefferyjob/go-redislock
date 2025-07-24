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
    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
    })

    // Create a context for canceling lock operations
    ctx := context.Background()

    // Create a RedisLock object
    lock := redislock.New(redisClient, "test_key")

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
    Lock() error
    // SpinLock Spinlock
    SpinLock(timeout time.Duration) error
    // UnLock Unlocking
    UnLock() error
    // Renew Manual renewal
    Renew() error
    
    // FairLock Fair lock locking
    FairLock(requestId string) error
    // SpinFairLock Spin Fair Lock
    SpinFairLock(requestId string, timeout time.Duration) error
    // FairUnLock Fair Lock Unlock
    FairUnLock(requestId string) error
    // FairRenew Fair Lock Renew
    FairRenew(requestId string) error
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


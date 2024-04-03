# go-redislock

[![Go](https://img.shields.io/badge/Go->=1.18-green)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/jefferyjob/go-redislock.svg)](https://github.com/jefferyjob/go-redislock/releases)
[![Action](https://github.com/jefferyjob/go-redislock/workflows/Go/badge.svg?branch=main)](https://github.com/jefferyjob/go-redislock/actions)
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

The SDK is already compatible with Redis clients, so the 'go redis/Redis' package you introduced needs to be greater than or equal to 8.

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-redis/redis/v8"
    redislock "github.com/jefferyjob/go-redislock"
)

func main() {
    // Create a Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    // Create a context for canceling lock operations
    ctx := context.Background()

    // Create a RedisLock object
    lock := redislock.New(ctx, redisClient, "test_key")

    // acquire lock
    err := lock.Lock()
    if err != nil {
        fmt.Println("lock acquisition failed：", err)
        return
    }
    defer lock.UnLock() // unlock

    // Perform tasks during lockdown
    // ...

    fmt.Println("task execution completed")
}
```

## Advanced usage

### Custom Token
You can customize the lock's token functionality by using the `WithToken` option:
```go
lock := redislock. New(ctx, redisClient,
    redislock. WithToken("your token")
)
```

### Custom lock timeout
You can customize the lock's timeout function by using the `WithTimeout` option:
```go
lock := redislock. New(ctx, redisClient,
    redislock.WithTimeout(time.Duration(10) * time.Second)
)
```

### Automatic renewal
You can enable automatic renewal when acquiring a lock by using the `WithAutoRenew` option:
```go
lock := redislock.New(ctx, redisClient,
	redislock.WithAutoRenew(true)
)
```

When using auto-renew, locks are automatically renewed periodically after acquisition to prevent locks from expiring. To manually renew a lock, the `Renew` method can be called.

### Spin lock
A spinlock is a way to repeatedly acquire a lock until it becomes available. You can use the `SpinLock` method to implement a spinlock:
```go
err := lock.SpinLock(time.Duration(5) * time.Second) // Try to acquire the lock, wait up to 5 seconds
if err != nil {
    fmt.Println("spinlock timeout：", err)
    return
}
defer lock.UnLock() // unlock

// Perform tasks during lockdown
// ...
```

### Lock release
If you wish to release the lock manually without waiting for the lock to timeout, you can use the `UnLock` method:
```go
err := lock.Lock()
if err != nil {
    fmt.Println("lock acquisition failed：", err)
    return
}

// perform tasks

err = lock.UnLock() // Release the lock manually
if err != nil {
    fmt.Println("lock release failed：", err)
    return
}
```

## Precautions
- Please make sure your Redis server is set up correctly, connected and running properly.
- When using the automatic renewal function, ensure that there is no long-term blocking during task execution, so as not to cause the renewal to fail.
- Consider using appropriate timeout settings to avoid deadlocks due to network issues etc.
- Try to ensure that the same key is used to acquire and release locks to ensure correctness.
- In the process of using locks, it is recommended to carefully design and test critical code blocks to avoid race conditions and deadlock problems.

## License
This library is licensed under the MIT. See the LICENSE file for details.


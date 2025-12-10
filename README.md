# go-redislock

[![Go](https://img.shields.io/badge/Go->=1.21-green)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/jefferyjob/go-redislock.svg)](https://github.com/jefferyjob/go-redislock/releases)
[![Action](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml/badge.svg)](https://github.com/jefferyjob/go-redislock/actions/workflows/go.yml)
[![Report](https://goreportcard.com/badge/github.com/jefferyjob/go-redislock)](https://goreportcard.com/report/github.com/jefferyjob/go-redislock)
[![Coverage](https://codecov.io/gh/jefferyjob/go-redislock/branch/main/graph/badge.svg)](https://codecov.io/gh/jefferyjob/go-redislock)
[![Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jefferyjob/go-redislock)
[![License](https://img.shields.io/github/license/jefferyjob/go-redislock)](https://github.com/jefferyjob/go-redislock/blob/main/LICENSE)

[English](README.en.md) | 简体中文

## 介绍
go-redislock 是一个用于 Go 的库，用于使用 Redis 作为后端存储提供分布式锁功能。确保在分布式环境中的并发访问下实现数据共享和资源互斥。我们的分布式锁具备可靠性、高性能、超时机制、可重入性和灵活的锁释放方式等特性，简化了分布式锁的使用，让您专注于业务逻辑的实现。

我们实现了以下关键能力：

- 🔒 普通分布式锁（可重入）
- 🔁 自旋锁
- ⚖️ 公平锁（FIFO 顺序）
- 🧵读锁（多个读者并发访问，互斥写者）
- ✍️写锁（独占访问资源）
- 🔄 手动续期与自动续期
- ✅ 多 Redis 客户端适配（v7/v8/v9、go-zero）

## 快速开始

### 安装
```bash
go get -u github.com/jefferyjob/go-redislock
```

### 使用Demo
```go
package main

import (
	"context"
	"fmt"

	redislock "github.com/jefferyjob/go-redislock"
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/V9"
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

### 配置选项
| **选项函数**                        | **说明**           | **默认值** |
| ----------------------------------- |------------------|---------|
| WithTimeout(d time.Duration)        | 锁超时时间（TTL）       | 5s      |
| WithAutoRenew()                     | 是否自动续期           | false   |
| WithToken(token string)             | 可重入锁 Token（唯一标识） | 随机 UUID |
| WithRequestTimeout(d time.Duration) | 公平锁队列最大等待时间      | 同 TTL   |


## 核心功能一览
### 普通锁
| 方法名                        | 说明                   |
|------------------------------|------------------------|
| `Lock(ctx)`                  | 获取普通锁（支持可重入）   |
| `SpinLock(ctx, timeout)`     | 自旋方式获取普通锁        |
| `UnLock(ctx)`                | 解锁操作                |
| `Renew(ctx)`                 | 手动续期                |

### 公平锁（FIFO）
| 方法名                                      | 说明                 |
|--------------------------------------------|----------------------|
| `FairLock(ctx, requestId)`                 | 获取公平锁（FIFO）      |
| `SpinFairLock(ctx, requestId, timeout)`    | 自旋方式获取公平锁      |
| `FairUnLock(ctx, requestId)`               | 公平锁解锁            |
| `FairRenew(ctx, requestId)`                | 公平锁续期            |

### 读锁
| 方法名                      | 说明          |
|--------------------------|-------------|
| `RLock(ctx)`             | 获取读锁（支持可重入） |
| `SpinRLock(ctx, timeout)` | 自旋方式获取读锁    |
| `UnLRock(ctx)`            | 解锁操作        |
| `RRenew(ctx)`             | 手动续期        |

### 写锁
| 方法名                      | 说明          |
|--------------------------|-------------|
| `WLock(ctx)`             | 获取写锁（支持可重入） |
| `SpinWLock(ctx, timeout)` | 自旋方式获取写锁    |
| `UnWLock(ctx)`            | 解锁操作        |
| `WRenew(ctx)`             | 手动续期        |

### 接口定义如下
```go
type RedisLockInter interface {
	// Lock 加锁
	Lock(ctx context.Context) error
	// SpinLock 自旋锁
	SpinLock(ctx context.Context, timeout time.Duration) error
	// UnLock 解锁
	UnLock(ctx context.Context) error
	// Renew 手动续期
	Renew(ctx context.Context) error

	// FairLock 公平锁加锁
	FairLock(ctx context.Context, requestId string) error
	// SpinFairLock 自旋公平锁
	SpinFairLock(ctx context.Context, requestId string, timeout time.Duration) error
	// FairUnLock 公平锁解锁
	FairUnLock(ctx context.Context, requestId string) error
	// FairRenew 公平锁续期
	FairRenew(ctx context.Context, requestId string) error

    // RLock 读锁加锁
    RLock(ctx context.Context) error
    // RUnLock 读锁解锁
    RUnLock(ctx context.Context) error
    // SpinRLock 自旋读锁
    SpinRLock(ctx context.Context, timeout time.Duration) error
    // RRenew 读锁续期
    RRenew(ctx context.Context) error
    
    // WLock 写锁加锁
    WLock(ctx context.Context) error
    // WUnLock 写锁解锁
    WUnLock(ctx context.Context) error
    // SpinWLock 自旋写锁
    SpinWLock(ctx context.Context, timeout time.Duration) error
    // WRenew 写锁续期
    WRenew(ctx context.Context) error
}
```

## Redis客户端支持
go-redislock 提供高度可扩展的客户端适配机制，已内置支持以下主流 Redis 客户端，详细示例请参考 [examples](examples) 。

| Redis客户端版本       | 包路径                                                      | 是否支持 |
|------------------|----------------------------------------------------------| -------- |
| go-redis v7      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V7` | ✅        |
| go-redis v8      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V8` | ✅        | 
| go-redis v9      | `github.com/jefferyjob/go-redislock/adapter/go-redis/V9` | ✅        | 
| go-zero redis    | `github.com/jefferyjob/go-redislock/adapter/go-zero/V1`  | ✅        | 

如您使用的 Redis 客户端不在上述列表中，也可以实现接口 `RedisInter` 来接入任意 Redis 客户端。


## 注意事项
- 每次加锁建议使用新的锁实例。
- 加锁和解锁必须使用同一个 key 和 token。
- 默认 TTL 是 5 秒，建议根据任务耗时自行设置。
- 自动续期适合无阻塞任务，避免长时间阻塞。
- 建议关键逻辑中使用 `defer unlock`，防止泄露。
- 建议对锁获取失败、重试等行为做日志或监控。
- 公平锁需传入唯一的 requestId（建议使用 UUID）。
- 读锁可并发，写锁互斥，避免读写冲突。
- 联锁中任一子锁失败，会释放已加成功的锁。
- Redis 不可用时可能造成死锁风险。

## 许可证
本库采用 MIT 进行授权。有关详细信息，请参阅 LICENSE 文件。


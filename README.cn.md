# go-redislock

[![Go](https://img.shields.io/badge/Go->=1.24-green)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/jefferyjob/go-redislock.svg)](https://github.com/jefferyjob/go-redislock/releases)
[![Action](https://github.com/jefferyjob/go-redislock/workflows/Go/badge.svg?branch=main)](https://github.com/jefferyjob/go-redislock/actions)
[![Report](https://goreportcard.com/badge/github.com/jefferyjob/go-redislock)](https://goreportcard.com/report/github.com/jefferyjob/go-redislock)
[![Coverage](https://codecov.io/gh/jefferyjob/go-redislock/branch/main/graph/badge.svg)](https://codecov.io/gh/jefferyjob/go-redislock)
[![Doc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jefferyjob/go-redislock)
[![License](https://img.shields.io/github/license/jefferyjob/go-redislock)](https://github.com/jefferyjob/go-redislock/blob/main/LICENSE)

[English](README.md) | 简体中文

## 介绍
go-redislock 是一个用于 Go 的库，用于使用 Redis 作为后端存储提供分布式锁功能。确保在分布式环境中的并发访问下实现数据共享和资源互斥。我们的分布式锁具备可靠性、高性能、超时机制、可重入性和灵活的锁释放方式等特性，简化了分布式锁的使用，让您专注于业务逻辑的实现。

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
	"github.com/redis/go-redis/v9"
)

func main() {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建 Redis 客户端适配器
	// 注意：根据不同的 Redis 客户端包使用不同的适配器
	rdbAdapter := redislock.NewRedisV9Adapter(rdb)

	// 创建用于取消锁定操作的上下文
	ctx := context.Background()

	// 创建 RedisLock 对象
	lock := redislock.New(rdbAdapter, "test_key")

	// 获取锁
	err := lock.Lock(ctx)
	if err != nil {
		fmt.Println("lock获取失败：", err)
		return
	}
	defer lock.UnLock(ctx) // 解锁

	// 锁定期间执行任务
	// ...
	fmt.Println("任务执行完成")
}
```

### 配置选项
| **选项函数**                        | **说明**           | **默认值** |
| ----------------------------------- |------------------|---------|
| WithTimeout(d time.Duration)        | 锁超时时间（TTL）       | 5s      |
| WithAutoRenew()                     | 是否自动续期           | false   |
| WithToken(token string)             | 可重入锁 Token（唯一标识） | 随机 UUID |
| WithRequestTimeout(d time.Duration) | 公平锁队列最大等待时间      | 同 TTL   |

### API 速查
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
}
```


## 注意事项
- 每次加锁都创建一个新的 `RedisLock` 实例。
- 请确保您的 Redis 服务器设置正确，并且能够正常连接和运行。
- 在使用自动续约功能时，确保在任务执行期间没有出现长时间的阻塞，以免导致续约失败。
- 考虑使用适当的超时设置，以避免由于网络问题等原因导致死锁。
- 尽量保证使用相同的 key 来获取和释放锁，以确保正确性。
- 在使用锁的过程中，建议对关键代码块进行精心设计和测试，以避免出现竞态条件和死锁问题。

## 许可证
本库采用 MIT 进行授权。有关详细信息，请参阅 LICENSE 文件。


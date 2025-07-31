# 写法推荐
下面是关于「每次加锁都创建一个新的 RedisLock 实例」的 Good Code（推荐写法） 与 Bad Code（反例），并说明其中的设计差异和原因。

## 代码示例
```go
package main

import (
	"context"
	redislock "github.com/jefferyjob/go-redislock"
)

// GoodLock Good Code（推荐）
// 每次加锁创建新实例，隔离上下文、参数与状态，线程安全，可复用 lock 组件。
//
// 特点：
//
//	每次调用 AcquireLock 都会创建一个新的 RedisLock 实例；
//	状态隔离，不会共享字段（如上下文、key、token）；
//	可用于并发场景；
//	更清晰的生命周期控制；
//	避免共享对象带来的副作用。
func GoodLock(ctx context.Context, rdb redislock.RedisInter) error {
	lock := redislock.New(rdb, "test_key")
	err := lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer lock.UnLock(ctx)

	// 业务逻辑...

	return nil
}

// Bad: 复用一个全局锁对象，易发生状态覆盖和并发冲突
var globalLock = &redislock.RedisLock{
	// redis: 提前初始化好的 redis 实例
}

// BadLock Bad Code（不推荐）
// 重复复用 RedisLock 实例，容易状态污染、竞态问题。
//
// 问题
//
//	使用全局实例，存在 共享状态污染问题；
//	多个并发请求时，lockKey、lockToken、Context 等字段会互相覆盖；
//	不利于单元测试（无法注入依赖）；
//	不利于多租户、异步逻辑；
//	隐藏了状态变化，代码可读性差，容易产生隐藏 bug。
func BadLock(ctx context.Context) error {
	err := globalLock.Lock(ctx)
	if err != nil {
		return err
	}
	defer globalLock.UnLock(ctx)

	// 业务逻辑...

	return nil
}
```


## 总结建议
| 项目       | 每次创建新实例（Good） | 重用全局实例（Bad） |
| -------- | ------------- | ----------- |
| 并发安全性    | ✅ 高           | ❌ 低         |
| 状态隔离     | ✅ 是           | ❌ 否         |
| 单测注入依赖   | ✅ 容易          | ❌ 困难        |
| 可维护性     | ✅ 强           | ❌ 弱         |
| 多协程支持    | ✅ 支持          | ❌ 易错        |
| 错误排查方便程度 | ✅ 高           | ❌ 低         |

如果你使用的是 `sync.Pool` 优化性能，也应该创建新的 `RedisLock` 实例，再从池中取结构体字段（如脚本、锁逻辑等）重用，而不是锁对象本身共享。



# 推荐写法：每次加锁创建新的 RedisLock 实例
在使用 Redis 实现分布式锁时，**强烈推荐每次加锁都创建一个新的 `RedisLock` 实例**。这样可以有效避免状态污染和并发冲突，提高系统的健壮性与可维护性。

## ✅ 推荐写法（Good Code）

```go
package main

import (
	"context"
	redislock "github.com/jefferyjob/go-redislock"
)

// GoodLock 每次加锁创建新的 RedisLock 实例。
// 优点：上下文隔离、线程安全、生命周期清晰，适用于高并发和异步场景。
func GoodLock(ctx context.Context, rdb redislock.RedisInter) error {
	lock := redislock.New(rdb, "test_key")
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.UnLock(ctx)

	// 执行业务逻辑...
	return nil
}
```

### ✅ 优势说明：
* 每次调用 `New()` 创建新实例，**不共享内部状态**；
* **上下文（context）、锁 key、token 等字段独立**，无交叉影响；
* **支持并发调用与异步逻辑**，线程安全；
* **生命周期更清晰**，避免资源复用带来的隐藏副作用；
* 易于单元测试，**可注入 Redis 依赖、Mock 接口**。

---

## ❌ 不推荐写法（Bad Code）
```go
// 不推荐：复用全局 RedisLock 实例，存在并发安全问题
var globalLock = &redislock.RedisLock{
	// redis: 提前初始化的 Redis 实例
}

func BadLock(ctx context.Context) error {
	if err := globalLock.Lock(ctx); err != nil {
		return err
	}
	defer globalLock.UnLock(ctx)

	// 执行业务逻辑...
	return nil
}
```

### ❌ 问题分析：
* **共享状态**：多个调用会复用同一个 `RedisLock` 实例，导致上下文、key、token 相互覆盖；
* **并发冲突**：在高并发场景下，容易出现竞态条件、死锁或误释放；
* **可测性差**：全局对象难以注入替代依赖，不利于单元测试与 Mock；
* **可读性差**：锁的生命周期和状态变化隐藏在共享实例中，排查问题困难；
* **不适合异步/多租户逻辑**：实例中字段可能被并发修改，引发不可预期的问题。

---

## 🔍 推荐与反例对比总结
| 项目          | 每次创建新实例 ✅（推荐） | 重用全局实例 ❌（不推荐） |
| ----------- | ------------- | ------------- |
| 并发安全性       | ✅ 高           | ❌ 低           |
| 状态隔离        | ✅ 完全隔离        | ❌ 共享易污染       |
| 单元测试支持      | ✅ 易于注入依赖      | ❌ 不利于测试       |
| 可读性与维护性     | ✅ 生命周期清晰      | ❌ 状态隐蔽难排查     |
| 异步/多协程兼容性   | ✅ 安全支持        | ❌ 易出错         |
| 适配多场景（如多租户） | ✅ 灵活支持        | ❌ 难以拓展        |


## 🚀 性能优化建议：结合 sync.Pool 使用
如果你使用的是 `sync.Pool` 优化性能，也应该创建新的 `RedisLock` 实例，再从池中取结构体字段（如脚本、锁逻辑等）重用，而不是锁对象本身共享。


## ✅ 总结建议
始终遵循以下原则：

* 每次加锁都 **新建 RedisLock 实例**；
* 不复用带有状态的锁对象；
* 用结构体字段缓存来提升性能，而不是锁实例本身；
* 保持锁的 **短生命周期、强隔离、可测试、可维护**。

这样可大幅提升分布式锁的健壮性与开发效率，避免由状态复用带来的复杂问题。
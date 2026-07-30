# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概览

go-redislock 是一个基于 Redis 的 Go 分布式锁库。核心代码在仓库根目录（`lock*.go`），锁的原子性语义由 `lua/` 下的 Lua 脚本通过 Redis `EVAL` 保证，脚本经 `//go:embed` 编译进二进制。

### 多模块结构（重要）

本仓库是 **Go 多模块仓库**（不是单 module，且 `go.work` 被 `.gitignore` 忽略）。每个目录有独立的 `go.mod`：

- 根模块：`github.com/jefferyjob/go-redislock` —— 锁核心实现，对外发布的主包。
- `adapter/go-redis/V7`、`adapter/go-redis/V8`、`adapter/go-redis/V9`、`adapter/go-zero/V1` —— 各客户端适配器，均为独立 module，通过 `replace github.com/jefferyjob/go-redislock => ../../..` 指向根。
- `tests/` —— 独立 module，集成风格测试（依赖真实 Redis）。
- `examples/demo` —— 独立 module 的示例。

**关键后果**：在根目录执行 `go test ./...` 或 `go test ./adapter/...` **不会**递归进入各适配器子模块。修改适配器代码时，必须 `cd` 进对应目录单独运行 `go test`。

## 构建与测试命令

```bash
make build        # go build -v ./...
make test         # go test -v ./...   （仅根模块）
make lint         # golangci-lint run
make bench        # go test -benchmem -bench .
make cover        # 带竞态检测：go test -race -coverprofile=coverage.txt -covermode=atomic ./...
make mocks        # 基于 lock.go 重新生成 mock（mockgen -source=lock.go -destination=mocks/lock.go）
make run-redis    # 启动本地 Redis 容器，端口映射 63790:6379（redis:5.0.3-alpine）
```

运行单个测试：`go test -run TestWriteLock ./tests/`（或 `go test -run FuncName ./` 指定包路径）。

测试依赖真实 Redis（连接 `127.0.0.1:63790`）：先 `make run-redis` 或自备 Redis。CI 中使用 redis:7.2。

变更依赖后必须执行 `go mod tidy` —— CI 会检查工作区是否因此产生变更，若有则失败。

## 核心架构

### 接口分层

- **`RedisInter`**（`lock.go`）：Redis 客户端的最小能力抽象，只有一个方法 `Eval(ctx, script, keys, args) RedisCmd`。适配器实现此接口即可接入任意 Redis 客户端。
- **`RedisCmd`**：`Eval` 返回结果的抽象，提供 `Result()` 和 `Int64()`。
- **`RedisLockInter`**：完整的锁能力接口，按锁类型分组定义方法。
- **`RedisLock`**：实现 `RedisLockInter` 的具体类型。一个实例对应一把锁（一个 key+token）。**建议每次加锁都 `New` 一个新实例。**

### 配置选项（Option 模式）

`New(client, key, opts...)` 接受 functional options：

| Option | 含义 | 默认值 |
|---|---|---|
| `WithTimeout` | 锁 TTL | 5s（`lockTime`） |
| `WithAutoRenew` | 开启后台自动续期 | false |
| `WithToken` | 重入/所有权标识 | `lock_token:<uuid>` |
| `WithRequestTimeout` | 公平锁在队列中的最大等待时间 | 同 TTL（`requestTimeout`） |

### 锁类别与对应的 Lua 脚本

每种锁的 `*Lock.go` / `*UnLock.go` / `*Renew.go` 三个脚本配套使用，Go 侧通过 `//go:embed` 加载：

| 锁类型 | 入口方法 | 脚本文件 | Redis 数据结构 |
|---|---|---|---|
| 可重入锁（普通锁） | `Lock`/`SpinLock`/`UnLock`/`Renew` | `lua/reentrant*.lua` | String `{key}` + 计数器 `{key}:count:{token}` |
| 公平锁（FIFO） | `FairLock`/`SpinFairLock`/`FairUnLock`/`FairRenew` | `lua/fair*.lua` | String `{key}` + 有序集合队列 `{key}:queue` |
| 读锁 | `RLock`/`SpinRLock`/`RUnLock`/`RRenew` | `lua/read*.lua` | 单个 HASH，含 `mode`/`rcount`/`r:{owner}` |
| 写锁 | `WLock`/`SpinWLock`/`WUnLock`/`WRenew` | `lua/write*.lua` | 同一 HASH，含 `mode`/`writer`/`wcount` |

> 读写锁共用同一个 key（HASH），通过 `mode` 字段在 `read`/`write` 间切换；读锁支持并发重入，写锁独占。当只有当前请求者自己持有时，读锁可升级为写锁；持写锁者也可重入读锁。

### Lua 脚本与 Go 的契约

- **返回值约定**：所有脚本成功返回 `1`，失败返回 `0`。Go 侧 `Int64()` 取值后判断 `!= 1` 即失败。
- **错误处理**：Redis 调用报错时，Go 侧统一用 `errors.Join(err, ErrException)` 包装（见 `vars.go` 中的哨兵错误：`ErrLockFailed`、`ErrUnLockFailed`、`ErrSpinLockTimeout`、`ErrSpinLockDone`、`ErrLockRenewFailed`、`ErrException`）。
- **ARGV 顺序**：修改 Lua 脚本时必须同步检查 Go 侧 `Eval` 的 `args...` 参数顺序（见 `lock_*.go` 中的 `.Eval(...)` 调用），二者必须严格对应。
- **Redis Cluster 兼容**：可重入锁/公平锁用 hash tag `{KEYS[1]}` 包裹主 key，确保主锁与计数器/队列落在同一 slot。

### 自动续期与自旋

- `WithAutoRenew`：加锁成功后启动 goroutine，以 `lockTimeout/3` 为周期调用对应的 `Renew`；`UnLock`/`FairUnLock` 等会取消该 goroutine。
- 自旋锁（`SpinLock` 等）：在 `timeout` 内每 `100ms` 重试一次，期间响应 `ctx.Done()`。

## 关键约定与陷阱

- **`MultiLock` 系列方法（`MultiLock`/`MultiUnLock`/`SpinMultiLock`/`MultiRenew`）当前未实现**，调用即 `panic("implement me")`（见 `lock_multi.go`）；对应脚本 `lua/multi*.lua` 已存在但未接入。
- **方法命名**：读锁解锁是 `RUnLock`、写锁解锁是 `WUnLock`（以代码 `lock.go` 的接口为准，README 个别表格存在笔误）。
- 修改 Lua 脚本、适配器实现或锁语义时，应同步补充/更新 `tests/` 或 `adapter/*/` 下的测试，必要时更新 `examples/` 和文档。
- 文档与注释为中英双语（英文描述在前，中文在后）。锁相关标识符保持现有命名风格（`Lock`/`UnLock`/`Renew`/`FairLock`/`RLock`/`WLock`）。
- 代码风格：制表符缩进，提交前 `gofmt`，导出标识符用 PascalCase。变更依赖后务必 `go mod tidy`。

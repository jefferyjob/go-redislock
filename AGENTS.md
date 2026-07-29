# Repository Guidelines

## 项目结构与模块组织
核心锁实现位于仓库根目录，例如 `lock.go`、`lock_reentrant.go`、`lock_read.go`、`lock_write.go`。运行时依赖的 Redis Lua 脚本存放在 `lua/`。适配器位于 `adapter/` 下，并按客户端版本拆分为独立子模块，例如 `adapter/go-redis/V9`、`adapter/go-zero/V1`，每个子模块都有自己的 `go.mod` 和测试。集成风格测试位于 `tests/`，示例代码位于 `examples/`，补充说明文档位于 `docs/`。

## 构建、测试与开发命令
本项目使用 Go 1.21。

- `make build`：执行 `go build -v ./...`，构建全部包。
- `make test`：执行 `go test -v ./...`，运行主测试集。
- `make cover`：运行带竞态检测的测试，并生成 `coverage.txt`。
- `make lint`：执行 `golangci-lint run`。
- `make run-redis`：启动本地 Redis 容器，端口为 `localhost:63790`，供集成测试使用。
- `go test ./tests/...`：运行 `tests/` 下的外部调用方式测试。
- `go test ./adapter/...` 不会递归进入各适配器子模块；修改适配器代码时，请进入对应目录单独执行 `go test`。

## 代码风格与命名约定
遵循标准 Go 风格：使用制表符缩进，提交前运行 `gofmt`，导出标识符采用 PascalCase。锁相关类型和方法应保持现有命名风格一致，例如 `Lock`、`UnLock`、`Renew`、`FairLock`、`RLock`、`WLock`。尽量按锁行为拆分小文件。变更依赖后务必执行 `go mod tidy`；CI 会检查其是否导致工作区变更。

## 测试规范
测试基于 Go 内置 `testing`，`tests/` 中使用了 `testify`。测试文件命名为 `*_test.go`，测试函数名应清晰表达行为，例如 `TestWriteLock...`。凡是修改锁语义、适配器实现或 Lua 脚本，都应补充或更新测试。涉及 Redis 的测试请先执行 `make run-redis`。

## 提交与 Pull Request 规范
最近提交以简短、直接的英文主题为主，常见形式包括带作用域前缀的 `docs: readme update`，或聚焦改动内容的描述，例如 `Refactor test files and update Redis adapter import paths`。提交应保持单一主题、描述明确。PR 需要说明行为变化、列出受影响模块、注明 Redis 或测试影响，并在适用时关联 Issue。若变更涉及公开 API 或适配器路径，请同步更新示例或文档。

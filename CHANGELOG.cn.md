- codecov：测试分析 [#32](https://github.com/jefferyjob/go-redislock/pull/32)
- Go多版本CI测试 [#33](https://github.com/jefferyjob/go-redislock/pull/33)

## v1.1.2
- Dependabot 计划间隔每周 [#27](https://github.com/jefferyjob/go-redislock/pull/27)
- 删除毫无意义的 `sync.Mutex` [#28](https://github.com/jefferyjob/go-redislock/pull/28)
- 优化可重入锁的命名 [#29](https://github.com/jefferyjob/go-redislock/pull/29)
- 更新问题表单 [#31](https://github.com/jefferyjob/go-redislock/pull/31)

## v1.1.1
- 单元测试覆盖与错误优化 [#25](https://github.com/jefferyjob/go-redislock/pull/25)
- 错误修复：在并发情况下，token相似会导致多次获取锁 [#26](https://github.com/jefferyjob/go-redislock/pull/26)

## v1.1.0
- 兼容新版本`redis/go-redis` [#17](https://github.com/jefferyjob/go-redislock/pull/17)
- 错误统一定义 [#18](https://github.com/jefferyjob/go-redislock/pull/18)
- 删除未使用的选项方法 [#19](https://github.com/jefferyjob/go-redislock/pull/19)
- 调整自动续订时间 [#20](https://github.com/jefferyjob/go-redislock/pull/20)
- 将 `github.com/redis/go-redis/v9` 从 `9.5.4` 升级到 `9.6.1` [#23](https://github.com/jefferyjob/go-redislock/pull/23)

## v1.0.3
- 优化Lua脚本 [#16](https://github.com/jefferyjob/go-redislock/pull/16)

## v1.0.2
- 讲 `v1.0.0` 标记废弃 [#15](https://github.com/jefferyjob/go-redislock/pull/15)
- 将 `codecov/codecov-action` 升级到版本4 [#11](https://github.com/jefferyjob/go-redislock/pull/11)

## v1.0.1
- 修复包名问题 [#10](https://github.com/jefferyjob/go-redislock/pull/10)

## v1.0.0
- 利用 Redis 后端存储，确保分布式锁的稳定性和可靠性
- 提供简单易用的 API，轻松实现加锁、解锁、自旋锁、自动续期和手动续期等功能
- 支持自定义超时时间和自动续期，根据实际需求进行灵活配置
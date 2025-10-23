## v1.4.0
- Bump github.com/zeromicro/go-zero from 1.8.5 to 1.9.0 by @dependabot[bot] in [#85](https://github.com/jefferyjob/go-redislock/pull/85)
- 将 README 默认语言改为中文 by @jefferyjob in [#86](https://github.com/jefferyjob/go-redislock/pull/86)
- Bump github.com/stretchr/testify from 1.10.0 to 1.11.0 by @dependabot[bot] in [#87](https://github.com/jefferyjob/go-redislock/pull/87)
- 更新 changelog 文件 by @jefferyjob in [#88](https://github.com/jefferyjob/go-redislock/pull/88)
- 新增读锁和写锁功能 by @jefferyjob in [#89](https://github.com/jefferyjob/go-redislock/pull/89)
- 文档更新 by @jefferyjob in [#90](https://github.com/jefferyjob/go-redislock/pull/90)
- Bump actions/setup-go from 5 to 6 by @dependabot[bot] in [#91](https://github.com/jefferyjob/go-redislock/pull/91)
- Bump github.com/redis/go-redis/v9 from 9.12.1 to 9.13.0 by @dependabot[bot] in [#92](https://github.com/jefferyjob/go-redislock/pull/92)
- Bump github.com/redis/go-redis/v9 from 9.13.0 to 9.14.0 by @dependabot[bot] in [#96](https://github.com/jefferyjob/go-redislock/pull/96)
- Bump github.com/zeromicro/go-zero from 1.9.0 to 1.9.1 by @dependabot[bot] in [#97](https://github.com/jefferyjob/go-redislock/pull/97)
- 为读锁和写锁增加单元测试代码 by @jefferyjob in [#98](https://github.com/jefferyjob/go-redislock/pull/98)
- 因为 [v9.15.1](https://github.com/redis/go-redis/releases/tag/v9.15.1) 更新调整逻辑 by @jefferyjob in [#99](https://github.com/jefferyjob/go-redislock/pull/99)
- 新增与移除 Lua 脚本 by @jefferyjob in [#100](https://github.com/jefferyjob/go-redislock/pull/100)
- Bump github/codeql-action from 3 to 4 by @dependabot[bot] in [#103](https://github.com/jefferyjob/go-redislock/pull/103)
- Bump github.com/redis/go-redis/v9 from 9.14.0 to 9.14.1 by @dependabot[bot] in [#105](https://github.com/jefferyjob/go-redislock/pull/105)
- 文档优化 by @jefferyjob in [#107](https://github.com/jefferyjob/go-redislock/pull/107)

## v1.3.0
- ci actions 标签配置 [#57](https://github.com/jefferyjob/go-redislock/pull/57)
- 更新 change log [#58](https://github.com/jefferyjob/go-redislock/pull/58)
- 支持公平锁加锁、解锁、自旋锁 [#59](https://github.com/jefferyjob/go-redislock/pull/59)
- 新增示例 demo [#60](https://github.com/jefferyjob/go-redislock/pull/60)
- 文件名调整 [#61](https://github.com/jefferyjob/go-redislock/pull/61)
- 支持中英文注释 [#62](https://github.com/jefferyjob/go-redislock/pull/62)
- 优化 context 定义 [#63](https://github.com/jefferyjob/go-redislock/pull/63)
- 每个方法传入 ctx [#66](https://github.com/jefferyjob/go-redislock/pull/66)
- 支持适配不同的 Redis 客户端包 [#67](https://github.com/jefferyjob/go-redislock/pull/67)
- 增加 good code 和 bad code 示例 [#69](https://github.com/jefferyjob/go-redislock/pull/69)
- 完善 examples 示例内容 [#70](https://github.com/jefferyjob/go-redislock/pull/70)
- 本地单元测试 [#71](https://github.com/jefferyjob/go-redislock/pull/71)
- 优化 MustNewRedisAdapter [#72](https://github.com/jefferyjob/go-redislock/pull/72)
- 提交单元测试代码 [#73](https://github.com/jefferyjob/go-redislock/pull/73)
- 修复单元测试 panic 错误（无法收集） [#74](https://github.com/jefferyjob/go-redislock/pull/74)
- 优化 codecov.yml 配置 [#75](https://github.com/jefferyjob/go-redislock/pull/75)
- 移除 gozero 和 goframe 的 redis adapter [#76](https://github.com/jefferyjob/go-redislock/pull/76)
- 回滚 “移除 gozero 和 goframe 的 redis adapter” [#78](https://github.com/jefferyjob/go-redislock/pull/78)
- 优化文档内容 [#79](https://github.com/jefferyjob/go-redislock/pull/79)
- 新建 Adapter 包 [#80](https://github.com/jefferyjob/go-redislock/pull/80)
- CI 增加 redis 安装流程 [#81](https://github.com/jefferyjob/go-redislock/pull/81)
- 适配器单元测试 [#82](https://github.com/jefferyjob/go-redislock/pull/82)
- Bump github.com/redis/go-redis/v9 依赖：9.11.0 → 9.12.1 [#83](https://github.com/jefferyjob/go-redislock/pull/83)
- Bump actions/checkout from 3 → 5 [#84](https://github.com/jefferyjob/go-redislock/pull/84)

## v1.2.0
- go版本升级到1.24  [#54](https://github.com/jefferyjob/go-redislock/pull/54)

## v1.1.4
- 修复lua脚本支持集群哈希卡槽的错误 [#44](https://github.com/jefferyjob/go-redislock/pull/44)

## v1.1.3
- codecov：测试分析 [#32](https://github.com/jefferyjob/go-redislock/pull/32)
- Go多版本CI测试 [#33](https://github.com/jefferyjob/go-redislock/pull/33)
- 更新lua脚本为毫秒单位 [#35](https://github.com/jefferyjob/go-redislock/pull/35)
- 更新changelog文件 [#37](https://github.com/jefferyjob/go-redislock/pull/37)
- 修改文档中的错误 [#38](https://github.com/jefferyjob/go-redislock/pull/38)
- 修复可重入锁解锁 [#39](https://github.com/jefferyjob/go-redislock/pull/39)

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
## v1.1.4
- Script attempted to access a non local key in a cluster node script [#44](https://github.com/jefferyjob/go-redislock/pull/44)

## v1.1.3
- codecov:Test Analytics [#32](https://github.com/jefferyjob/go-redislock/pull/32)
- Go multi-version CI test [#33](https://github.com/jefferyjob/go-redislock/pull/33)
- feat:update ttl to ms  [#35](https://github.com/jefferyjob/go-redislock/pull/35)
- Update the changelog file [#37](https://github.com/jefferyjob/go-redislock/pull/37)
- fix: Modify errors in the document [#38](https://github.com/jefferyjob/go-redislock/pull/38)
- Fix reentrant lock unlock [#39](https://github.com/jefferyjob/go-redislock/pull/39)

## v1.1.2
- Dependabot scheduled every week [#27](https://github.com/jefferyjob/go-redislock/pull/27)
- Delete meaningless `sync.Mutex` [#28](https://github.com/jefferyjob/go-redislock/pull/28)
- Optimize the naming of reentrant locks [#29](https://github.com/jefferyjob/go-redislock/pull/29)
- Update to issue question form [#31](https://github.com/jefferyjob/go-redislock/pull/31)

## v1.1.1
- Unit test coverage and error optimization [#25](https://github.com/jefferyjob/go-redislock/pull/25)
- Fix: In concurrent situations, similar tokens will cause multiple lock acquisitions [#26](https://github.com/jefferyjob/go-redislock/pull/26)

## v1.1.0
- Compatible with new version `redis/go-redis` [#17](https://github.com/jefferyjob/go-redislock/pull/17)
- Unify error definitions [#18](https://github.com/jefferyjob/go-redislock/pull/18)
- Delete unused option methods [#19](https://github.com/jefferyjob/go-redislock/pull/19)
- Adjust auto-renewal time [#20](https://github.com/jefferyjob/go-redislock/pull/20)
- Upgrade `github.com/redis/go-redis/v9` from `9.5.4` to `9.6.1` [#23](https://github.com/jefferyjob/go-redislock/pull/23)

## v1.0.3
- Optimize Lua scripts [#16](https://github.com/jefferyjob/go-redislock/pull/16)

## v1.0.2
- Mark `v1.0.0` as deprecated [#15](https://github.com/jefferyjob/go-redislock/pull/15)
- Upgrade `codecov/codecov-action` to version 4 [#11](https://github.com/jefferyjob/go-redislock/pull/11)

## v1.0.1
- Fix package name issue [#10](https://github.com/jefferyjob/go-redislock/pull/10)

## v1.0.0
- Use Redis backend storage to ensure the stability and reliability of distributed locks
- Provides an easy-to-use API to easily implement functions such as lock, unlock, spin lock, automatic renewal, and manual renewal
- Support custom timeout and automatic renewal, flexible configuration according to actual needs
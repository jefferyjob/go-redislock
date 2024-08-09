# Version update record

## v1.0.0
- Use Redis backend storage to ensure the stability and reliability of distributed locks
- Provides an easy-to-use API to easily implement functions such as lock, unlock, spin lock, automatic renewal and manual renewal
- Support custom timeout and automatic renewal, flexible configuration according to actual needs

## v1.0.1
- Fix package name issue

## v1.0.2
- Mark `v1.0.0` as deprecated #15
- Upgrade `codecov/codecov-action` to version 4 #11

## v1.0.3
- Optimize Lua scripts #16

## v1.1.0
- Compatible with new version `redis/go-redis` #17
- Unify error definitions #18
- Delete unused option methods #19
- Adjust auto-renewal time #20
- Upgrade `github.com/redis/go-redis/v9` from `9.5.4` to `9.6.1` #23

## v1.1.1
- Unit test coverage and error optimization #25
- Fix: In concurrent situations, similar tokens will cause multiple lock acquisitions #26

## v1.1.2
- Dependabot scheduled every week #27
- Delete meaningless `sync.Mutex` #28
- Optimize the naming of reentrant locks #29
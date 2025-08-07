# Redis client adapter

go-redislock provides a highly scalable client adaptation mechanism, and has built-in support for the following mainstream Redis clients. For detailed examples, please refer to [examples](../examples/adapter) .

| Redis Client Version | Package path | Supported |
|------------------|---------------------------------------------------| -------- |
| go-redis v7      | `github.com/jefferyjob/go-redislock/adapter/v7`   | ✅        |
| go-redis v8      | `github.com/jefferyjob/go-redislock/adapter/v8`   | ✅        |
| go-redis v9      | `github.com/jefferyjob/go-redislock/adapter/v9`   | ✅        |
| go-zero redis    | `github.com/jefferyjob/go-redislock/adapter/gozero` | ✅        |
| goframe v1 redis | `github.com/jefferyjob/go-redislock/adapter/gf/v1` | ✅        |
| goframe v2 redis | `github.com/jefferyjob/go-redislock/adapter/gf/v2` | ✅        |

If the Redis client you are using is not in the above list, you can also implement the interface `RedisInter` to connect to any Redis client.

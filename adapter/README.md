# Redis client adapter
go-redislock provides a highly scalable client adaptation mechanism, and has built-in support for the following mainstream Redis clients. For detailed examples, please refer to [examples](../examples/adapter) .

If the Redis client you are using is not in the above list, you can also implement the interface `RedisInter` to connect to any Redis client.

## Import the adapter package
### go-redis
- v9
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v9"
)
```

- v8
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v8"
)
```

- v7
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/go-redis/v7"
)
```

### go-zero
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/go-zero"
)
```

### goframe
- v2
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/gf/v2"
)
```

- v1
```
import (
	adapter "github.com/jefferyjob/go-redislock/adapter/gf"
)
```
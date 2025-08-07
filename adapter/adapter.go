package adapter

import (
	"fmt"
	redislock "github.com/jefferyjob/go-redislock"
	adapterV9 "github.com/jefferyjob/go-redislock/adapter/v9"
	v9 "github.com/redis/go-redis/v9"
	"log"
	"runtime/debug"
)

func New(rawClient interface{}) (redislock.RedisInter, error) {
	switch client := rawClient.(type) {
	case *v9.Client:
		return adapterV9.New(client), nil
	default:
		return nil, fmt.Errorf("unsupported redis client type: %T", rawClient)
	}
}

func MustNew(rawClient interface{}) redislock.RedisInter {
	adapter, err := New(rawClient)
	if err != nil {
		msg := fmt.Sprintf("%+v\n\n%s", err.Error(), debug.Stack())
		log.Println(msg)
		panic(msg)
	}
	return adapter
}

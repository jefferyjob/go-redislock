package adapter

import (
	"fmt"
	"testing"

	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
	gfRdbV1 "github.com/gogf/gf/database/gredis"
	gfRdbV2 "github.com/gogf/gf/v2/database/gredis"
	v9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	zeroRdb "github.com/zeromicro/go-zero/core/stores/redis"
)

func TestNew(t *testing.T) {
	v7Client := &v7.Client{}
	v8Client := &v8.Client{}
	v9Client := &v9.Client{}
	zeroClient := &zeroRdb.Redis{}
	gfClient1 := &gfRdbV1.Redis{}
	gfClient2 := &gfRdbV2.Redis{}

	tests := []struct {
		name    string
		client  interface{}
		wantErr bool
	}{
		{"redis.v7", v7Client, false},
		{"redis.v8", v8Client, false},
		{"redis.v9", v9Client, false},
		{"go-zero redis", zeroClient, false},
		{"gf redis v1", gfClient1, false},
		{"gf redis v2", gfClient2, false},
		{"unsupported", "xxx", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := New(tt.client)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
			}
		})
	}
}

func TestMustNew(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Expected panic:", err)
		}
	}()

	MustNew("xxx")
}

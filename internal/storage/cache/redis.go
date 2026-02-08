package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "43.13.141.101:6379", // 确保你本地或服务器已安装 Redis
		Password: "jkesh1024",          // 没有密码则留空
		DB:       0,                    // 默认数据库
	})
}

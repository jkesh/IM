package cache

import (
	"IM/internal/config"
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

func InitRedis(cfg config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(Ctx).Err(); err != nil {
		return err
	}

	RDB = client
	return nil
}

func Available() bool {
	return RDB != nil
}

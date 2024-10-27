package cache

import (
	"context"

	"github.com/dk5761/go-serv/configs"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func InitRedisClient(cfg configs.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return rdb
}

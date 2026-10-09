package redisx

import (
	"context"
	"time"
	"webhook-relay/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisX struct {
	*redis.Client
}

func NewRedis(ctx context.Context, cfg *config.RedisConfig) (*RedisX, error) {
	redisCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   cfg.Db,
	})
	ping := client.Ping(redisCtx)
	if err := ping.Err(); err != nil {
		return nil, err
	}
	return &RedisX{
		client,
	}, nil
}

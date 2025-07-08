package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"weather-api/internal/config"
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}

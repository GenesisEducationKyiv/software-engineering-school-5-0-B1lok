package redis

import (
	"context"
	"time"

	"weather-service/internal/config"

	"github.com/redis/go-redis/v9"
)

const (
	readTimeout  = 2 * time.Second
	writeTimeout = 2 * time.Second
	dialTimeout  = 3 * time.Second
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DialTimeout:  dialTimeout,
	})

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}

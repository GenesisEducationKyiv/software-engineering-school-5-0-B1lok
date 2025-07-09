package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func Get[T any](ctx context.Context, client *redis.Client, key string) (*T, error) {
	raw, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("redis: error for key=%s: %v\n", key, err)
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		log.Printf("redis: unmarshal error for key=%s: %v\n", key, err)
		return nil, err
	}

	log.Printf("redis: get key=%s success\n", key)
	return &result, nil
}

func Set(
	ctx context.Context,
	client *redis.Client,
	key string,
	value any,
	ttl time.Duration,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := client.Set(ctx, key, data, ttl).Err(); err != nil {
		return err
	}

	log.Printf("redis: set key=%s ttl=%s success\n", key, ttl.String())
	return nil
}

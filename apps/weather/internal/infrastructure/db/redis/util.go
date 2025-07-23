package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/redis/go-redis/v9"
)

func Get[T any](ctx context.Context, client *redis.Client, key string) (*T, error) {
	raw, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to get value")
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to unmarshal value")
		return nil, err
	}

	log.Debug().Str("key", key).Msg("redis: get key success")
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

	log.Debug().Str("key", key).Msg("redis: set key success")
	return nil
}

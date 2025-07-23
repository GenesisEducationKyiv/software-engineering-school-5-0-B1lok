package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	appRedis "weather-service/internal/infrastructure/db/redis"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Validate(ctx context.Context, city string) (*string, error)
}

type MetricsRecorder interface {
	CacheHit()
	CacheMiss()
}

type ProxyClient struct {
	delegate Client
	redis    *redis.Client
	ttl      time.Duration
	prefix   string
	recorder MetricsRecorder
}

func NewProxyClient(
	delegate Client,
	redisClient *redis.Client,
	ttl time.Duration,
	prefix string,
	recorder MetricsRecorder,
) *ProxyClient {
	return &ProxyClient{
		delegate: delegate,
		redis:    redisClient,
		ttl:      ttl,
		prefix:   prefix,
		recorder: recorder,
	}
}

func (c *ProxyClient) Validate(ctx context.Context, city string) (*string, error) {
	key := fmt.Sprintf("%s:%s", c.prefix, strings.ToLower(city))

	cachedCity, err := appRedis.Get[string](ctx, c.redis, key)
	if err == nil {
		c.recorder.CacheHit()
		return cachedCity, nil
	}

	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	cityValidated, err := c.delegate.Validate(ctx, city)
	if err != nil {
		return nil, err
	}

	if cityValidated != nil {
		if err := appRedis.Set(ctx, c.redis, key, cityValidated, c.ttl); err != nil {
			log.Error().Err(err).Str("key", key).Msg("validator: failed to cache cityValidated")
		}
	}

	return cityValidated, nil
}

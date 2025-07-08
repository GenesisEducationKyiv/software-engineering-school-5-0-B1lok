package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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

	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cachedCity string
		if err := json.Unmarshal([]byte(cached), &cachedCity); err == nil {
			c.recorder.CacheHit()
			return &cachedCity, nil
		}
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}
	cityValidated, err := c.delegate.Validate(ctx, city)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cityValidated)
	if err == nil {
		_ = c.redis.Set(ctx, key, data, c.ttl).Err()
	}

	return cityValidated, nil
}

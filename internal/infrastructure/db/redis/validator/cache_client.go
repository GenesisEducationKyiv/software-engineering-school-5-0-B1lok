package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Validate(city string) (*string, error)
}

type ProxyClient struct {
	delegate Client
	redis    *redis.Client
	ttl      time.Duration
	prefix   string
}

func NewProxyClient(
	delegate Client,
	redisClient *redis.Client,
	ttl time.Duration,
	prefix string,
) *ProxyClient {
	return &ProxyClient{
		delegate: delegate,
		redis:    redisClient,
		ttl:      ttl,
		prefix:   prefix,
	}
}

func (c *ProxyClient) Validate(city string) (*string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", c.prefix, strings.ToLower(city))

	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cachedCity string
		if err := json.Unmarshal([]byte(cached), &cachedCity); err == nil {
			return &cachedCity, nil
		}
	}

	cityValidated, err := c.delegate.Validate(city)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cityValidated)
	if err == nil {
		_ = c.redis.Set(ctx, key, data, c.ttl).Err()
	}

	return cityValidated, nil
}

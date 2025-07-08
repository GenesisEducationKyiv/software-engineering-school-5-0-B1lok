package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"weather-api/internal/domain"
)

type Client interface {
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
}

type MetricsRecorder interface {
	CacheHit()
	CacheMiss()
}

type ProxyClient struct {
	delegate Client
	redis    *redis.Client
	provider TTLProvider
	prefix   string
	recorder MetricsRecorder
}

type ForecastType string

const (
	ForecastCurrent ForecastType = "current"
	ForecastHourly  ForecastType = "hourly"
	ForecastDaily   ForecastType = "daily"
)

type TTLProvider interface {
	TTL(forecastType ForecastType) time.Duration
}

func NewProxyClient(
	delegate Client,
	redisClient *redis.Client,
	provider TTLProvider,
	prefix string,
	recorder MetricsRecorder,
) *ProxyClient {
	return &ProxyClient{
		delegate: delegate,
		redis:    redisClient,
		provider: provider,
		prefix:   prefix,
		recorder: recorder,
	}
}

func (c *ProxyClient) GetDailyForecast(
	ctx context.Context, city string,
) (*domain.WeatherDaily, error) {
	key := fmt.Sprintf("%s:%s:%s", c.prefix, ForecastDaily, strings.ToLower(city))

	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cachedWeather WeatherDaily
		if err := json.Unmarshal([]byte(cached), &cachedWeather); err == nil {
			c.recorder.CacheHit()
			return ToDomainWeatherDaily(&cachedWeather), nil
		}
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetDailyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data, err := json.Marshal(ToDTOWeatherDaily(weather))
		if err == nil {
			_ = c.redis.Set(ctx, key, data, c.provider.TTL(ForecastDaily)).Err()
		}
	}

	return weather, nil
}

func (c *ProxyClient) GetHourlyForecast(
	ctx context.Context, city string,
) (*domain.WeatherHourly, error) {
	key := fmt.Sprintf("%s:%s:%s", c.prefix, ForecastHourly, strings.ToLower(city))

	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cachedWeather WeatherHourly
		if err := json.Unmarshal([]byte(cached), &cachedWeather); err == nil {
			c.recorder.CacheHit()
			return ToDomainWeatherHourly(&cachedWeather), nil
		}
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetHourlyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data, err := json.Marshal(ToDTOWeatherHourly(weather))
		if err == nil {
			_ = c.redis.Set(ctx, key, data, c.provider.TTL(ForecastHourly)).Err()
		}
	}

	return weather, nil
}

func (c *ProxyClient) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	key := fmt.Sprintf("%s:%s:%s", c.prefix, ForecastCurrent, strings.ToLower(city))

	cached, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		var cachedWeather Weather
		if err := json.Unmarshal([]byte(cached), &cachedWeather); err == nil {
			c.recorder.CacheHit()
			return ToDomainWeather(&cachedWeather), nil
		}
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data, err := json.Marshal(ToDTOWeather(weather))
		if err == nil {
			_ = c.redis.Set(ctx, key, data, c.provider.TTL(ForecastCurrent)).Err()
		}
	}

	return weather, nil
}

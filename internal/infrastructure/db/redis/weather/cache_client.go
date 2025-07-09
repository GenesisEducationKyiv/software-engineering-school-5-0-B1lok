package weather

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"weather-api/internal/domain"
	appRedis "weather-api/internal/infrastructure/db/redis"
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
	key := c.getForecastKey(ForecastDaily, city)

	cached, err := appRedis.Get[WeatherDaily](ctx, c.redis, key)
	if err == nil {
		c.recorder.CacheHit()
		return ToDomainWeatherDaily(cached), nil
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetDailyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data := ToDTOWeatherDaily(weather)
		if err := appRedis.Set(
			ctx, c.redis, key, data, c.provider.TTL(ForecastDaily),
		); err != nil {
			log.Printf("proxy: failed to set cache for key %s: %v\n", key, err)
		}
	}

	return weather, nil
}

func (c *ProxyClient) GetHourlyForecast(
	ctx context.Context, city string,
) (*domain.WeatherHourly, error) {
	key := c.getForecastKey(ForecastHourly, city)

	cached, err := appRedis.Get[WeatherHourly](ctx, c.redis, key)
	if err == nil {
		c.recorder.CacheHit()
		return ToDomainWeatherHourly(cached), nil
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetHourlyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data := ToDTOWeatherHourly(weather)
		if err := appRedis.Set(
			ctx, c.redis, key, data, c.provider.TTL(ForecastHourly),
		); err != nil {
			log.Printf("proxy: failed to set cache for key %s: %v\n", key, err)
		}
	}

	return weather, nil
}

func (c *ProxyClient) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	key := c.getForecastKey(ForecastCurrent, city)

	cached, err := appRedis.Get[Weather](ctx, c.redis, key)
	if err == nil {
		c.recorder.CacheHit()
		return ToDomainWeather(cached), nil
	}
	if errors.Is(err, redis.Nil) {
		c.recorder.CacheMiss()
	}

	weather, err := c.delegate.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}

	if weather != nil {
		data := ToDTOWeather(weather)
		if err := appRedis.Set(
			ctx, c.redis, key, data, c.provider.TTL(ForecastCurrent),
		); err != nil {
			log.Printf("proxy: failed to set cache for key %s: %v\n", key, err)
		}
	}

	return weather, nil
}

func (c *ProxyClient) getForecastKey(forecastType ForecastType, city string) string {
	return fmt.Sprintf("%s:%s:%s", c.prefix, forecastType, strings.ToLower(city))
}

package weather

import (
	"context"

	"weather-api/internal/domain"
)

type Client interface {
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
}

type Repository struct {
	client Client
}

func NewRepository(client Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	return r.client.GetWeather(ctx, city)
}

func (r *Repository) GetDailyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherDaily, error) {
	return r.client.GetDailyForecast(ctx, city)
}

func (r *Repository) GetHourlyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherHourly, error) {
	return r.client.GetHourlyForecast(ctx, city)
}

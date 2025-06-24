package weather

import (
	"context"

	"weather-api/internal/domain"
)

type Handler interface {
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
	SetNext(next Handler)
}

type Repository struct {
	handler Handler
}

func NewRepository(provider Handler) *Repository {
	return &Repository{handler: provider}
}

func (r *Repository) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	return r.handler.GetWeather(ctx, city)
}

func (r *Repository) GetDailyForecast(
	ctx context.Context, city string,
) (*domain.WeatherDaily, error) {
	return r.handler.GetDailyForecast(ctx, city)
}

func (r *Repository) GetHourlyForecast(
	ctx context.Context, city string,
) (*domain.WeatherHourly, error) {
	return r.handler.GetHourlyForecast(ctx, city)
}

package weather

import (
	"weather-api/internal/domain"
)

type Client interface {
	GetDailyForecast(city string) (*domain.WeatherDaily, error)
	GetHourlyForecast(city string) (*domain.WeatherHourly, error)
	GetWeather(city string) (*domain.Weather, error)
}

type Repository struct {
	client Client
}

func NewRepository(client Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) GetWeather(city string) (*domain.Weather, error) {
	return r.client.GetWeather(city)
}

func (r *Repository) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	return r.client.GetDailyForecast(city)
}

func (r *Repository) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	return r.client.GetHourlyForecast(city)
}

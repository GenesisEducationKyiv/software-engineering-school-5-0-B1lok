package weather

import (
	"context"

	"weather-service/internal/application/query"
	"weather-service/internal/domain"
)

type Reader interface {
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
}

type Service struct {
	reader Reader
}

func NewService(weatherRepository Reader) *Service {
	return &Service{reader: weatherRepository}
}

func (s *Service) GetWeather(ctx context.Context, city string) (*query.WeatherResult, error) {
	weather, err := s.reader.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}
	return toNewWeatherResult(weather), nil
}

func (s *Service) GetDailyForecast(
	ctx context.Context,
	city string,
) (*query.WeatherDailyResult, error) {
	weather, err := s.reader.GetDailyForecast(ctx, city)
	if err != nil {
		return nil, err
	}
	return toNewWeatherDailyResult(weather), nil
}

func (s *Service) GetHourlyForecast(
	ctx context.Context,
	city string,
) (*query.WeatherHourlyResult, error) {
	weather, err := s.reader.GetHourlyForecast(ctx, city)
	if err != nil {
		return nil, err
	}
	return toNewWeatherHourlyResult(weather), nil
}

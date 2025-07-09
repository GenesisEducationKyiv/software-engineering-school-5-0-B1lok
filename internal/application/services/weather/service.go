package weather

import (
	"context"

	"weather-api/internal/application/common"
	"weather-api/internal/application/query"
	"weather-api/internal/domain"
)

type Reader interface {
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
}

type Service struct {
	weatherRepository Reader
}

func NewService(weatherRepository Reader) *Service {
	return &Service{weatherRepository: weatherRepository}
}

func (s *Service) GetWeather(
	ctx context.Context, city string,
) (*query.WeatherQueryResult, error) {
	weather, err := s.weatherRepository.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}
	var queryResult query.WeatherQueryResult
	queryResult.Result = toNewWeatherResult(weather)
	return &queryResult, nil
}

func toNewWeatherResult(weather *domain.Weather) *common.WeatherResult {
	return &common.WeatherResult{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
}

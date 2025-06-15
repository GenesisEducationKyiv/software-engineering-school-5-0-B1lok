package services

import (
	"context"

	"weather-api/internal/application/common"
	"weather-api/internal/application/query"
	"weather-api/internal/domain"
)

type WeatherReader interface {
	GetWeather(ctx context.Context, city string) (*domain.Weather, error)
}

type WeatherService struct {
	weatherRepository WeatherReader
}

func NewWeatherService(weatherRepository WeatherReader) *WeatherService {
	return &WeatherService{weatherRepository: weatherRepository}
}

func (s *WeatherService) GetWeather(
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

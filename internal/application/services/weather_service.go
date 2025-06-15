package services

import (
	"context"

	"weather-api/internal/application/common"
	"weather-api/internal/domain/models"

	"weather-api/internal/application/query"
	"weather-api/internal/domain/repositories"
)

type WeatherService struct {
	weatherRepository repositories.WeatherRepository
}

func NewWeatherService(weatherRepository repositories.WeatherRepository) *WeatherService {
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

func toNewWeatherResult(weather *models.Weather) *common.WeatherResult {
	return &common.WeatherResult{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
}

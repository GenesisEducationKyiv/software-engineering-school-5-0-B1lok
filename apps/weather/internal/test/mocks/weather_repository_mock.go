package mocks

import (
	"context"

	"weather-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockWeatherRepository struct {
	mock.Mock
}

func (m *MockWeatherRepository) GetWeather(
	ctx context.Context,
	city string,
) (*domain.Weather, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Weather), args.Error(1)
}

func (m *MockWeatherRepository) GetDailyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherDaily, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WeatherDaily), args.Error(1)
}

func (m *MockWeatherRepository) GetHourlyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherHourly, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WeatherHourly), args.Error(1)
}

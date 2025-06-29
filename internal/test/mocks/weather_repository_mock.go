package mocks

import (
	"weather-api/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockWeatherRepository struct {
	mock.Mock
}

func (m *MockWeatherRepository) GetWeather(city string) (*domain.Weather, error) {
	args := m.Called(city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Weather), args.Error(1)
}

func (m *MockWeatherRepository) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	args := m.Called(city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WeatherDaily), args.Error(1)
}

func (m *MockWeatherRepository) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	args := m.Called(city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WeatherHourly), args.Error(1)
}

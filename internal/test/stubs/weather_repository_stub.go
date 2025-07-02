package stubs

import (
	"sync"

	"weather-api/internal/domain"
	internalErrors "weather-api/internal/errors"
	pkgErrors "weather-api/pkg/errors"
)

type WeatherRepositoryStub struct {
	callCount           map[string]int
	mu                  sync.RWMutex
	GetWeatherFn        func(city string) (*domain.Weather, error)
	GetDailyForecastFn  func(city string) (*domain.WeatherDaily, error)
	GetHourlyForecastFn func(city string) (*domain.WeatherHourly, error)
}

func NewWeatherRepositoryStub() *WeatherRepositoryStub {
	return &WeatherRepositoryStub{
		GetWeatherFn:        nil,
		GetDailyForecastFn:  nil,
		GetHourlyForecastFn: nil,
		callCount:           make(map[string]int),
	}
}

func (s *WeatherRepositoryStub) GetWeather(city string) (*domain.Weather, error) {
	s.mu.Lock()
	s.callCount[city]++
	s.mu.Unlock()
	if s.GetWeatherFn != nil {
		return s.GetWeatherFn(city)
	}
	if city == "InvalidCity" {
		return nil, pkgErrors.New(internalErrors.ErrNotFound, "City not found")
	}
	return &domain.Weather{
		Temperature: 20.5,
		Humidity:    60.0,
		Description: "Clear sky",
	}, nil
}

func (s *WeatherRepositoryStub) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	if s.GetDailyForecastFn != nil {
		return s.GetDailyForecastFn(city)
	}
	return &domain.WeatherDaily{
		Location:   city,
		Date:       "2025-05-18",
		MaxTempC:   22.0,
		MinTempC:   15.0,
		AvgTempC:   18.5,
		WillItRain: false,
		ChanceRain: 10,
		WillItSnow: false,
		ChanceSnow: 0,
		Condition:  "Sunny",
		Icon:       "sunny.png",
	}, nil
}

func (s *WeatherRepositoryStub) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	if s.GetHourlyForecastFn != nil {
		return s.GetHourlyForecastFn(city)
	}
	return &domain.WeatherHourly{
		Location:   city,
		Time:       "12:00",
		TempC:      20.0,
		WillItRain: false,
		ChanceRain: 5,
		WillItSnow: false,
		ChanceSnow: 0,
		Condition:  "Partly cloudy",
		Icon:       "cloudy.png",
	}, nil
}

func (s *WeatherRepositoryStub) GetCallCount(city string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.callCount[city]
}

func (s *WeatherRepositoryStub) ResetCallCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCount = make(map[string]int)
}

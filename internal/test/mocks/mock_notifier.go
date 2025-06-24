package mocks

import (
	"github.com/stretchr/testify/mock"

	"weather-api/internal/domain"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) NotifyDailyWeather(
	subscription *domain.Subscription, weatherDaily *domain.WeatherDaily,
) error {
	args := m.Called(subscription, weatherDaily)
	return args.Error(0)
}

func (m *MockNotifier) NotifyHourlyWeather(
	subscription *domain.Subscription, weatherHourly *domain.WeatherHourly,
) error {
	args := m.Called(subscription, weatherHourly)
	return args.Error(0)
}

func (m *MockNotifier) NotifyConfirmation(
	subscription *domain.Subscription,
) error {
	args := m.Called(subscription)
	return args.Error(0)
}

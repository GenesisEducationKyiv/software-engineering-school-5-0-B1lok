//go:build unit
// +build unit

package weather

import (
	"context"
	"github.com/stretchr/testify/mock"
	"testing"

	"weather-api/internal/domain"
	internalErrors "weather-api/internal/errors"
	"weather-api/internal/test/mocks"
	pkgErrors "weather-api/pkg/errors"

	"github.com/stretchr/testify/assert"
)

const (
	validatedCity = "Berlin"
)

func TestWeatherService_GetWeather_Success(t *testing.T) {
	mockRepo := new(mocks.MockWeatherRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	mockWeather := &domain.Weather{
		Temperature: 20.5,
		Humidity:    70,
		Description: "Partly Cloudy",
	}

	mockRepo.On("GetWeather", mock.Anything, validatedCity).Return(mockWeather, nil)

	result, err := service.GetWeather(ctx, validatedCity)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockWeather.Temperature, result.Result.Temperature)
	assert.Equal(t, mockWeather.Humidity, result.Result.Humidity)
	assert.Equal(t, mockWeather.Description, result.Result.Description)
	mockRepo.AssertExpectations(t)
}

func TestWeatherService_GetWeather_EmptyCity(t *testing.T) {
	mockRepo := new(mocks.MockWeatherRepository)
	service := NewService(mockRepo)
	city := "invalid city123"
	expectedErr := pkgErrors.New(internalErrors.ErrInvalidInput, "invalid city123")
	ctx := context.Background()

	mockRepo.On("GetWeather", mock.Anything, city).Return(nil, expectedErr)

	result, err := service.GetWeather(ctx, city)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

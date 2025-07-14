//go:build unit
// +build unit

package weather_api

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
	appHttp "weather-service/internal/infrastructure/http"
)

type MockClock struct{}

func (MockClock) Now() time.Time {
	t, _ := time.Parse("2006-01-02 15:04", "2025-05-18 00:00")
	return t
}

func loadJSONFile(t *testing.T, filename string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path) // #nosec G304 -- filename is controlled and safe
	require.NoError(t, err, "Failed to read test data file: %s", path)
	return string(content)
}

func TestGetWeather(t *testing.T) {
	mockResponse := loadJSONFile(t, "current_weather_response.json")

	repo := NewClient(
		"http://mocked-weather-api.com",
		"dummy-http-server-key",
		appHttp.NoOpLogger{},
		MockClock{},
	)
	repo.client = appHttp.MockHTTPClient(appHttp.MockResponse{Body: mockResponse, StatusCode: http.StatusOK})
	ctx := context.Background()

	weather, err := repo.GetWeather(ctx, "London")

	require.NoError(t, err)
	require.NotNil(t, weather)
	assert.Equal(t, 11.0, weather.Temperature)
	assert.Equal(t, "Partly cloudy", weather.Description)
}

func TestGetDailyForecast(t *testing.T) {
	mockResponse := loadJSONFile(t, "forecast_response.json")

	repo := NewClient(
		"http://mocked-weather-api.com",
		"dummy-http-server-key",
		appHttp.NoOpLogger{},
		MockClock{},
	)
	repo.client = appHttp.MockHTTPClient(appHttp.MockResponse{Body: mockResponse, StatusCode: http.StatusOK})
	ctx := context.Background()

	forecast, err := repo.GetDailyForecast(ctx, "London")

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, "London", forecast.Location)
	assert.Equal(t, "Partly Cloudy", forecast.Condition)
	assert.Equal(t, 12.3, forecast.AvgTempC)
}

func TestGetHourlyForecast(t *testing.T) {
	mockResponse := loadJSONFile(t, "forecast_response.json")

	repo := NewClient("http://mocked-weather-api.com",
		"dummy-http-server-key",
		appHttp.NoOpLogger{},
		MockClock{},
	)
	repo.client = appHttp.MockHTTPClient(appHttp.MockResponse{Body: mockResponse, StatusCode: http.StatusOK})
	ctx := context.Background()

	forecast, err := repo.GetHourlyForecast(ctx, "London")

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, "London", forecast.Location)
	assert.Equal(t, "Clear", forecast.Condition)
	assert.Equal(t, 10.2, forecast.TempC)
}

func TestHandleAPIErrorResponse(t *testing.T) {
	mockErrorResponse := loadJSONFile(t, "not_found_response.json")

	repo := NewClient(
		"http://mocked-weather-api.com",
		"dummy-http-server-key",
		appHttp.NoOpLogger{},
		MockClock{},
	)
	repo.client = appHttp.MockHTTPClient(appHttp.MockResponse{Body: mockErrorResponse, StatusCode: http.StatusBadRequest})
	ctx := context.Background()

	_, err := repo.GetWeather(ctx, "NonExistentCity")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "City not found")
}

func TestDirectMapping(t *testing.T) {
	currentJSON := loadJSONFile(t, "current_weather_response.json")
	dailyJSON := loadJSONFile(t, "forecast_response.json")
	hourlyJSON := loadJSONFile(t, "forecast_response.json")

	var currentResponse WeatherRepositoryResponse
	var dailyResponse WeatherDailyResponse
	var hourlyResponse WeatherHourlyResponse

	err := json.Unmarshal([]byte(currentJSON), &currentResponse)
	require.NoError(t, err)

	err = json.Unmarshal([]byte(dailyJSON), &dailyResponse)
	require.NoError(t, err)

	err = json.Unmarshal([]byte(hourlyJSON), &hourlyResponse)
	require.NoError(t, err)

	weather := toWeather(&currentResponse)
	dailyForecast := toWeatherDaily(&dailyResponse)
	hourlyForecast := toWeatherHourly(&hourlyResponse, MockClock{}.Now())

	assert.NotNil(t, weather)
	assert.NotNil(t, dailyForecast)
	assert.NotNil(t, hourlyForecast)
}

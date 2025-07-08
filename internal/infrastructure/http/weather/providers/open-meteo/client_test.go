//go:build unit
// +build unit

package open_meteo

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
	appHttp "weather-api/internal/infrastructure/http"
)

var (
	openMeteoURL = "http://open-meteo-mock/v1"
	geoCodingURL = "http://geocoding-mock/v1"
)

type MockClock struct{}

func (MockClock) Now() time.Time {
	t, _ := time.Parse("2006-01-02T15:04", "2025-06-26T00:00")
	return t
}

func loadJSONFile(t *testing.T, filename string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path) // #nosec G304 -- filename is controlled and safe
	require.NoError(t, err, "Failed to read test data file: %s", path)
	return string(content)
}

func TestGetWeather(t *testing.T) {
	coordinatesResponse := loadJSONFile(t, "coordinates_response.json")
	currentResponse := loadJSONFile(t, "current_weather_response.json")

	mockResponses := map[string]appHttp.MockResponse{
		geoCodingURL: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoURL: {
			Body:       currentResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoURL, geoCodingURL, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)
	ctx := context.Background()

	weather, err := repo.GetWeather(ctx, "Kyiv")

	require.NoError(t, err)
	require.NotNil(t, weather)
	assert.Equal(t, 20.7, weather.Temperature)
	assert.Equal(t, "Overcast", weather.Description)
}

func TestGetDailyForecast(t *testing.T) {
	coordinatesResponse := loadJSONFile(t, "coordinates_response.json")
	dailyResponse := loadJSONFile(t, "daily_weather_response.json")
	location := "Kyiv"

	mockResponses := map[string]appHttp.MockResponse{
		geoCodingURL: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoURL: {
			Body:       dailyResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoURL, geoCodingURL, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)
	ctx := context.Background()

	forecast, err := repo.GetDailyForecast(ctx, location)

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, location, forecast.Location)
	assert.Equal(t, "Overcast", forecast.Condition)
	assert.Equal(t, 17.75, forecast.AvgTempC)
}

func TestGetHourlyForecast(t *testing.T) {
	coordinatesResponse := loadJSONFile(t, "coordinates_response.json")
	hourlyResponse := loadJSONFile(t, "hourly_weather_response.json")
	location := "Kyiv"

	mockResponses := map[string]appHttp.MockResponse{
		geoCodingURL: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoURL: {
			Body:       hourlyResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoURL, geoCodingURL, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)
	repo.SetClock(MockClock{})
	ctx := context.Background()

	forecast, err := repo.GetHourlyForecast(ctx, location)

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, location, forecast.Location)
	assert.Equal(t, "Overcast", forecast.Condition)
	assert.Equal(t, 18.0, forecast.TempC)
}

//go:build unit
// +build unit

package open_meteo

import (
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
	openMeteoUrl = "http://open-meteo-mock/v1"
	geoCodingUrl = "http://geocoding-mock/v1"
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
		geoCodingUrl: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoUrl: {
			Body:       currentResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoUrl, geoCodingUrl, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)

	weather, err := repo.GetWeather("Kyiv")

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
		geoCodingUrl: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoUrl: {
			Body:       dailyResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoUrl, geoCodingUrl, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)

	forecast, err := repo.GetDailyForecast(location)

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, location, forecast.Location)
	assert.Equal(t, "Overcast", forecast.Condition)
	assert.Equal(t, 17.75, forecast.AvgTempC)
}

func TestGetHourlyForecast(t *testing.T) {
	coordinatesResponse := loadJSONFile(t, "coordinates_response.json")
	dailyResponse := loadJSONFile(t, "hourly_weather_response.json")
	location := "Kyiv"

	mockResponses := map[string]appHttp.MockResponse{
		geoCodingUrl: {
			Body:       coordinatesResponse,
			StatusCode: http.StatusOK,
		},
		openMeteoUrl: {
			Body:       dailyResponse,
			StatusCode: http.StatusOK,
		},
	}

	repo := NewClient(openMeteoUrl, geoCodingUrl, appHttp.NoOpLogger{})
	repo.client = appHttp.MockHTTPClientWithResponses(mockResponses)
	repo.SetClock(MockClock{})

	forecast, err := repo.GetHourlyForecast(location)

	require.NoError(t, err)
	require.NotNil(t, forecast)

	assert.Equal(t, location, forecast.Location)
	assert.Equal(t, "Overcast", forecast.Condition)
	assert.Equal(t, 18.0, forecast.TempC)
}

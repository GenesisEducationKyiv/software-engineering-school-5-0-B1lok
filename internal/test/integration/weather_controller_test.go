//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"weather-api/internal/application/services/weather"
	cacheClient "weather-api/internal/infrastructure/db/redis/weather"
	cacheWeather "weather-api/internal/infrastructure/db/redis/weather/providers/weather-api"
	"weather-api/internal/infrastructure/prometheus"
	"weather-api/internal/interface/rest"
	"weather-api/internal/test/containers"
	"weather-api/internal/test/stubs"
	"weather-api/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type WeatherControllerTestSuite struct {
	suite.Suite
	Router         *gin.Engine
	WeatherRepo    *stubs.WeatherRepositoryStub
	RedisContainer *containers.RedisContainer
}

func (suite *WeatherControllerTestSuite) SetupSuite() {
	ctx := context.Background()

	redisContainer, err := containers.SetupRedisContainer(ctx)
	suite.Require().NoError(err)
	suite.RedisContainer = redisContainer

	weatherMetrics := prometheus.NewCacheMetrics("weather-api", "weather")

	weatherRepo := stubs.NewWeatherRepositoryStub()
	suite.WeatherRepo = weatherRepo

	cachedWeatherApiClient := cacheClient.NewProxyClient(
		weatherRepo,
		redisContainer.Client,
		cacheWeather.NewTTLProvider(),
		"weather-api",
		weatherMetrics,
	)
	weatherService := weather.NewService(cachedWeatherApiClient)
	weatherController := rest.NewWeatherController(weatherService)

	router := gin.Default()
	router.Use(middleware.ErrorHandler())
	api := router.Group("/api")
	api.GET("/weather", weatherController.GetWeather)

	suite.Router = router
}

func (suite *WeatherControllerTestSuite) SetupTest() {
	suite.RedisContainer.Client.FlushAll(context.Background())
	suite.WeatherRepo.ResetCallCount()
}

func (suite *WeatherControllerTestSuite) TestGetWeather() {
	req, reqErr := http.NewRequest(http.MethodGet, "/api/weather?city=London", nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()

	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Contains(response, "description")
	suite.Contains(response, "temperature")
	suite.Contains(response, "humidity")
}

func (suite *WeatherControllerTestSuite) TestGetWeatherInvalidCity() {
	req, reqErr := http.NewRequest(http.MethodGet, "/api/weather?city=InvalidCity", nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()

	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusNotFound, resp.Code)
}

func (suite *WeatherControllerTestSuite) TestGetWeatherInvalidQueryParam() {
	req, reqErr := http.NewRequest(http.MethodGet, "/api/weather?city=", nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()

	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusBadRequest, resp.Code)
}

func (suite *WeatherControllerTestSuite) TestCacheHitAndMiss() {
	city := "London"
	cacheKey := fmt.Sprintf("weather-api:%s:%s", "current", strings.ToLower(city))

	req1, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/weather?city=%s", city), nil)
	resp1 := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp1, req1)

	suite.Equal(http.StatusOK, resp1.Code)
	suite.Equal(1, suite.WeatherRepo.GetCallCount(city), "Repository should be called once on cache miss")

	val, err := suite.RedisContainer.Client.Get(context.Background(), cacheKey).Result()
	suite.Require().NoError(err)
	suite.NotEmpty(val)

	req2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/weather?city=%s", city), nil)
	resp2 := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp2, req2)

	suite.Equal(http.StatusOK, resp2.Code)
	suite.Equal(1, suite.WeatherRepo.GetCallCount(city), "Repository should not be called on cache hit")

	var response1, response2 map[string]interface{}

	err1 := json.Unmarshal(resp1.Body.Bytes(), &response1)
	suite.Require().NoError(err1, "Failed to unmarshal first response")

	err2 := json.Unmarshal(resp2.Body.Bytes(), &response2)
	suite.Require().NoError(err2, "Failed to unmarshal second response")

	suite.Equal(response1, response2)
}

func TestWeatherControllerTestSuite(t *testing.T) {
	suite.Run(t, new(WeatherControllerTestSuite))
}

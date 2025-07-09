//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"weather-api/internal/application/services/weather"

	"weather-api/internal/interface/rest"

	"weather-api/internal/test/stubs"
	"weather-api/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type WeatherControllerTestSuite struct {
	suite.Suite
	Router *gin.Engine
}

func (suite *WeatherControllerTestSuite) SetupSuite() {
	weatherRepo := stubs.NewWeatherRepositoryStub()
	weatherService := weather.NewService(weatherRepo)
	weatherController := rest.NewWeatherController(weatherService)

	router := gin.Default()
	router.Use(middleware.ErrorHandler())
	api := router.Group("/api")
	api.GET("/weather", weatherController.GetWeather)

	suite.Router = router
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

func TestWeatherControllerTestSuite(t *testing.T) {
	suite.Run(t, new(WeatherControllerTestSuite))
}

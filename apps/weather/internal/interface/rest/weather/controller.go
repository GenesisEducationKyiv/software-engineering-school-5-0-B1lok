package weather

import (
	"context"
	"net/http"

	"weather-service/internal/application/query"
	internalErrors "weather-service/internal/errors"
	"weather-service/internal/interface/rest/weather/dto/mapper"
	pkgErrors "weather-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetWeather(ctx context.Context, city string) (*query.WeatherResult, error)
	GetDailyForecast(ctx context.Context, city string) (*query.WeatherDailyResult, error)
	GetHourlyForecast(ctx context.Context, city string) (*query.WeatherHourlyResult, error)
}

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{
		service: service,
	}
}

func (h *Controller) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.service.GetWeather(ctx, city)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapper.ToWeatherResponse(weather))
}

func (h *Controller) GetDailyForecast(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.service.GetDailyForecast(ctx, city)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapper.ToWeatherDailyResponse(weather))
}

func (h *Controller) GetHourlyForecast(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.service.GetHourlyForecast(ctx, city)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapper.ToWeatherHourlyResponse(weather))
}

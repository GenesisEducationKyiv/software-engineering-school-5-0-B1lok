package rest

import (
	"context"
	"net/http"

	"weather-api/internal/application/query"
	"weather-api/internal/interface/rest/dto/mapper"

	"weather-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (*query.WeatherQueryResult, error)
}

type WeatherController struct {
	service WeatherService
}

func NewWeatherController(service WeatherService) *WeatherController {
	return &WeatherController{
		service: service,
	}
}

func (h *WeatherController) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(errors.New("Invalid request", http.StatusBadRequest)) //nolint:errcheck
		return
	}
	weather, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapper.ToWeatherResponse(weather.Result))
}

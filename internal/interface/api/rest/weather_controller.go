package rest

import (
	"net/http"

	"weather-api/internal/application/interfaces"
	"weather-api/internal/interface/api/rest/dto/mapper"
	"weather-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

type WeatherController struct {
	service interfaces.WeatherService
}

func NewWeatherController(service interfaces.WeatherService) *WeatherController {
	return &WeatherController{
		service: service,
	}
}

func (h *WeatherController) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		_ = c.Error(errors.New("Invalid request", http.StatusBadRequest))
		return
	}
	weather, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, mapper.ToWeatherResponse(weather.Result))
}

package weather

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/errs"
)

type Controller struct {
	client WeatherServiceClient
}

func NewController(client WeatherServiceClient) *Controller {
	return &Controller{client: client}
}

func (h *Controller) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.client.GetCurrentWeather(ctx, &CityRequest{City: city})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapCurrentWeather(weather))
}

func (h *Controller) GetDailyForecast(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.client.GetDailyWeather(ctx, &CityRequest{City: city})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapDailyWeather(weather))
}

func (h *Controller) GetHourlyForecast(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	weather, err := h.client.GetHourlyWeather(ctx, &CityRequest{City: city})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, mapHourlyWeather(weather))
}

package validator

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

type CityValidator interface {
	Validate(ctx context.Context, city string) (*string, error)
}

type Controller struct {
	validator CityValidator
}

func NewController(service CityValidator) *Controller {
	return &Controller{
		validator: service,
	}
}

func (h *Controller) ValidateCity(c *gin.Context) {
	cityQuery := c.Query("city")
	if cityQuery == "" {
		c.Error(pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")) //nolint:errcheck
		return
	}

	ctx := c.Request.Context()
	validatedCity, err := h.validator.Validate(ctx, cityQuery)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}
	if validatedCity == nil {
		c.Error(pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid request")) //nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, gin.H{"city": *validatedCity})
}

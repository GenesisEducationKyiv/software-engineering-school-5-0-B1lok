package subscription

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gateway/internal/errs"
)

type Controller struct {
	client SubscriptionServiceClient
}

type SubscribeRequestDto struct {
	Email     string `form:"email" binding:"required,email"`
	City      string `form:"city" binding:"required"`
	Frequency string `form:"frequency" binding:"required,oneof=hourly daily"`
}

func NewController(client SubscriptionServiceClient) *Controller {
	return &Controller{client: client}
}

func (ctr *Controller) Subscribe(c *gin.Context) {
	var req SubscribeRequestDto
	if err := c.ShouldBind(&req); err != nil {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid input")) //nolint:errcheck
		return
	}

	resp, err := ctr.client.Subscribe(c.Request.Context(), &SubscribeRequest{
		Email:     req.Email,
		City:      req.City,
		Frequency: req.Frequency,
	})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}

func (ctr *Controller) Confirm(c *gin.Context) {
	token := c.Param("token")
	if strings.TrimSpace(token) == "" {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid token")) //nolint:errcheck
		return
	}

	resp, err := ctr.client.Confirm(c.Request.Context(), &TokenRequest{Token: token})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}

func (ctr *Controller) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if strings.TrimSpace(token) == "" {
		c.Error(errs.NewHTTPError(http.StatusBadRequest, "Invalid token")) //nolint:errcheck
		return
	}

	resp, err := ctr.client.Unsubscribe(c.Request.Context(), &TokenRequest{Token: token})
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": resp.Message})
}

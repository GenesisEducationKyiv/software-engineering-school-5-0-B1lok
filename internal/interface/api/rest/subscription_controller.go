package rest

import (
	"net/http"
	"strings"

	"weather-api/internal/application/interfaces"
	"weather-api/internal/interface/api/rest/dto/request"
	"weather-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

type SubscriptionController struct {
	service interfaces.SubscriptionService
}

func NewSubscriptionController(service interfaces.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{service: service}
}

func (s *SubscriptionController) Subscribe(c *gin.Context) {
	var req request.SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(errors.New("Invalid input", http.StatusBadRequest)) //nolint:errcheck
		return
	}

	err := s.service.Subscribe(c.Request.Context(), req.ToSubscribeCommand())
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": "Subscription successful. Confirmation email sent."})
}

func (s *SubscriptionController) Confirm(c *gin.Context) {
	token := c.Param("token")
	if strings.TrimSpace(token) == "" {
		c.Error(errors.New("Invalid token", http.StatusBadRequest)) //nolint:errcheck
		return
	}

	err := s.service.Confirm(c.Request.Context(), token)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
}

func (s *SubscriptionController) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if strings.TrimSpace(token) == "" {
		c.Error(errors.New("Invalid token", http.StatusBadRequest)) //nolint:errcheck
		return
	}

	err := s.service.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		c.Error(err) //nolint:errcheck
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

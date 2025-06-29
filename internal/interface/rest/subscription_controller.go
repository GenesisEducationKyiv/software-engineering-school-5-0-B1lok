package rest

import (
	"context"
	"net/http"
	"strings"

	"weather-api/internal/application/command"
	"weather-api/internal/interface/rest/dto/request"

	"weather-api/pkg/errors"

	"github.com/gin-gonic/gin"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, subscribeCommand *command.SubscribeCommand) error
	Confirm(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
}

type SubscriptionController struct {
	service SubscriptionService
}

func NewSubscriptionController(service SubscriptionService) *SubscriptionController {
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

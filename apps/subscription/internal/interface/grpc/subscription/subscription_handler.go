package subscription

import (
	"context"
	"strings"

	"subscription-service/internal/application/command"
	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
)

type Service interface {
	Subscribe(ctx context.Context, subscribeCommand *command.SubscribeCommand) error
	Confirm(ctx context.Context, token string) error
	Unsubscribe(ctx context.Context, token string) error
}

type Handler struct {
	UnimplementedSubscriptionServiceServer
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (s *Handler) Subscribe(
	ctx context.Context,
	request *SubscribeRequest,
) (*ResponseMessage, error) {
	cmd := &command.SubscribeCommand{
		Email:     request.Email,
		City:      request.City,
		Frequency: request.Frequency,
	}
	err := s.service.Subscribe(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &ResponseMessage{Message: "Subscription successful. Confirmation email sent."}, nil
}

func (s *Handler) Confirm(ctx context.Context, request *TokenRequest) (*ResponseMessage, error) {
	token := strings.TrimSpace(request.Token)
	if token == "" {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid token")
	}
	err := s.service.Confirm(ctx, token)
	if err != nil {
		return nil, err
	}
	return &ResponseMessage{Message: "Subscription confirmed successfully"}, nil
}

func (s *Handler) Unsubscribe(
	ctx context.Context,
	request *TokenRequest,
) (*ResponseMessage, error) {
	token := strings.TrimSpace(request.Token)
	if token == "" {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid token")
	}
	err := s.service.Unsubscribe(ctx, token)
	if err != nil {
		return nil, err
	}
	return &ResponseMessage{Message: "Unsubscribed successfully"}, nil
}

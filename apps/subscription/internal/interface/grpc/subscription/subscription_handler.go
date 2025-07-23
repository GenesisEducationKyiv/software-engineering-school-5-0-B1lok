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
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	cmd := &command.SubscribeCommand{
		Email:     request.Email,
		City:      request.City,
		Frequency: request.Frequency,
	}
	err = s.service.Subscribe(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &ResponseMessage{Message: "Subscription successful. Confirmation email sent."}, nil
}

func (s *Handler) Confirm(ctx context.Context, request *TokenRequest) (*ResponseMessage, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	token := strings.TrimSpace(request.Token)

	err = s.service.Confirm(ctx, token)
	if err != nil {
		return nil, err
	}
	return &ResponseMessage{Message: "Subscription confirmed successfully"}, nil
}

func (s *Handler) Unsubscribe(
	ctx context.Context,
	request *TokenRequest,
) (*ResponseMessage, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	token := strings.TrimSpace(request.Token)

	err = s.service.Unsubscribe(ctx, token)
	if err != nil {
		return nil, err
	}
	return &ResponseMessage{Message: "Unsubscribed successfully"}, nil
}

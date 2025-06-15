package services

import (
	"context"
	"net/http"

	"weather-api/internal/application/command"
	"weather-api/internal/application/email"
	"weather-api/internal/domain"
	"weather-api/pkg/errors"
)

type ConfirmationSender interface {
	ConfirmationEmail(email *email.ConfirmationEmail) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error)
	ExistByLookup(ctx context.Context, lookup *domain.SubscriptionLookup) (bool, error)
	Update(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error)
	Delete(ctx context.Context, id uint) error
	FindByToken(ctx context.Context, token string) (*domain.Subscription, error)
}

type CityValidator interface {
	Validate(city string) (*string, error)
}

type SubscriptionService struct {
	repository SubscriptionRepository
	validator  CityValidator
	sender     ConfirmationSender
	host       string
}

func NewSubscriptionService(
	repository SubscriptionRepository,
	validator CityValidator,
	sender ConfirmationSender, host string,
) *SubscriptionService {
	return &SubscriptionService{
		repository: repository, validator: validator, sender: sender, host: host,
	}
}

func (s *SubscriptionService) Subscribe(
	ctx context.Context, subscribeCommand *command.SubscribeCommand,
) error {
	validatedCity, err := s.validator.Validate(subscribeCommand.City)
	if err != nil {
		return err
	}
	subscribeCommand.City = *validatedCity
	exists, err := s.repository.ExistByLookup(ctx, subscribeCommand.ToSubscriptionLookup())
	if err != nil {
		return errors.Wrap(
			err, "failed to check if email exists", http.StatusInternalServerError,
		)
	}
	if exists {
		return errors.New("Email already subscribed", http.StatusConflict)
	}

	newSubscription, err := domain.NewSubscription(
		subscribeCommand.Email,
		subscribeCommand.City,
		domain.Frequency(subscribeCommand.Frequency),
	)
	if err != nil {
		return errors.Wrap(err, "Invalid input", http.StatusBadRequest)
	}

	savedSubscription, err := s.repository.Create(ctx, newSubscription)
	if err != nil {
		return errors.Wrap(err, "failed to create subscription", http.StatusInternalServerError)
	}
	confirmationEmail := &email.ConfirmationEmail{
		To:        savedSubscription.Email,
		City:      savedSubscription.City,
		Frequency: string(savedSubscription.Frequency),
		Url:       s.host + "api/confirm/" + newSubscription.Token,
	}
	if err := s.sender.ConfirmationEmail(confirmationEmail); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) Confirm(ctx context.Context, token string) error {
	subscription, err := s.repository.FindByToken(ctx, token)
	if err != nil {
		return errors.Wrap(err, "failed to find subscription", http.StatusInternalServerError)
	}

	if subscription == nil {
		return errors.New("Token not found", http.StatusNotFound)
	}

	subscription.SetConfirmed(true)

	_, err = s.repository.Update(ctx, subscription)
	if err != nil {
		return errors.Wrap(err, "failed to update subscription", http.StatusInternalServerError)
	}
	return err
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, token string) error {
	subscription, err := s.repository.FindByToken(ctx, token)
	if err != nil {
		return errors.Wrap(err, "failed to find subscription", http.StatusInternalServerError)
	}

	if subscription == nil {
		return errors.New("Token not found", http.StatusNotFound)
	}

	return s.repository.Delete(ctx, subscription.ID)
}

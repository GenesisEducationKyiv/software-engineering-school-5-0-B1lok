package subscription

import (
	"context"

	"subscription-service/internal/application/event"
	"subscription-service/internal/application/event/subscription"

	"subscription-service/internal/application/command"
	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, event event.Event) error
}

type Repository interface {
	Create(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error)
	ExistByLookup(ctx context.Context, lookup *domain.SubscriptionLookup) (bool, error)
	Update(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error)
	Delete(ctx context.Context, id uint) error
	FindByToken(ctx context.Context, token string) (*domain.Subscription, error)
}

type CityValidator interface {
	Validate(ctx context.Context, city string) (*string, error)
}

type Service struct {
	repository Repository
	validator  CityValidator
	dispatcher EventDispatcher
}

func NewService(
	repository Repository,
	validator CityValidator,
	dispatcher EventDispatcher,
) *Service {
	return &Service{
		repository: repository, validator: validator, dispatcher: dispatcher,
	}
}

func (s *Service) Subscribe(
	ctx context.Context, subscribeCommand *command.SubscribeCommand,
) error {
	if err := s.setValidatedCity(ctx, subscribeCommand); err != nil {
		return err
	}
	exists, err := s.repository.ExistByLookup(ctx, subscribeCommand.ToSubscriptionLookup())
	if err != nil {
		return err
	}
	if exists {
		return pkgErrors.New(internalErrors.ErrConflict, "Email already subscribed")
	}

	newSubscription, err := s.createSubscription(ctx, subscribeCommand)
	if err != nil {
		return err
	}
	if err := s.dispatcher.Dispatch(ctx, &subscription.UserSubscribedEvent{
		Email:     newSubscription.Email,
		City:      newSubscription.City,
		Frequency: newSubscription.Frequency,
		Token:     newSubscription.Token,
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) Confirm(ctx context.Context, token string) error {
	sub, err := s.repository.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if sub == nil {
		return pkgErrors.New(internalErrors.ErrNotFound, "Token not found")
	}

	sub.SetConfirmed(true)

	_, err = s.repository.Update(ctx, sub)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) Unsubscribe(ctx context.Context, token string) error {
	sub, err := s.repository.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if sub == nil {
		return pkgErrors.New(internalErrors.ErrNotFound, "Token not found")
	}

	return s.repository.Delete(ctx, sub.ID)
}

func (s *Service) setValidatedCity(
	ctx context.Context,
	subscribeCommand *command.SubscribeCommand,
) error {
	validatedCity, err := s.validator.Validate(ctx, subscribeCommand.City)
	if err != nil {
		return err
	}
	subscribeCommand.City = *validatedCity
	return nil
}

func (s *Service) createSubscription(
	ctx context.Context,
	subscribeCommand *command.SubscribeCommand,
) (*domain.Subscription, error) {
	newSubscription, err := domain.NewSubscription(
		subscribeCommand.Email,
		subscribeCommand.City,
		domain.Frequency(subscribeCommand.Frequency),
	)
	if err != nil {
		return nil, err
	}

	savedSubscription, err := s.repository.Create(ctx, newSubscription)
	if err != nil {
		return nil, err
	}

	return savedSubscription, nil
}

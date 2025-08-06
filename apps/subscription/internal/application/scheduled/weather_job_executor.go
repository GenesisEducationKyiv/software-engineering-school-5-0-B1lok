package scheduled

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"subscription-service/internal/application/event"
	"subscription-service/internal/application/event/subscription"

	"subscription-service/internal/domain"
)

type SubscriptionReader interface {
	StreamSubscriptions(
		ctx context.Context,
		frequency *domain.Frequency,
	) (<-chan domain.Subscription, <-chan error, error)
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event event.Event) error
}

type WeatherJobExecutor struct {
	subscriptionReader SubscriptionReader
	frequency          domain.Frequency
	dispatcher         EventDispatcher
}

func NewWeatherJobExecutor(
	subscriptionReader SubscriptionReader,
	frequency domain.Frequency,
	dispatcher EventDispatcher,
) *WeatherJobExecutor {
	return &WeatherJobExecutor{
		subscriptionReader: subscriptionReader,
		frequency:          frequency,
		dispatcher:         dispatcher,
	}
}

func (e *WeatherJobExecutor) Execute(ctx context.Context) error {
	log.Info().Str("frequency", string(e.frequency)).Msg("Weather job started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	err := e.processSubscriptionsStream(ctx, &e.frequency)
	if err != nil {
		return err
	}

	log.Info().Str("frequency", string(e.frequency)).Msg("Weather job completed successfully")
	return nil
}

func (e *WeatherJobExecutor) processSubscriptionsStream(
	ctx context.Context,
	frequency *domain.Frequency,
) error {
	subCh, errCh, err := e.subscriptionReader.StreamSubscriptions(ctx, frequency)
	if err != nil {
		return err
	}

	if err = e.processChannels(ctx, subCh, errCh); err != nil {
		return err
	}

	return nil
}

func (e *WeatherJobExecutor) processChannels(
	ctx context.Context,
	subCh <-chan domain.Subscription,
	errCh <-chan error,
) error {
	hasErrorHappened := false

	for subCh != nil || errCh != nil {
		select {
		case <-ctx.Done():
			log.Info().Err(ctx.Err()).Msg("context cancelled")
			return ctx.Err()
		case sub, ok := <-subCh:
			if !ok {
				subCh = nil
				continue
			}
			if err := e.handleSubscription(ctx, &sub); err != nil {
				log.Error().Err(err).Msg("dispatching weather update")
				hasErrorHappened = true
			}
		case streamErr, ok := <-errCh:
			if !ok {
				errCh = nil
				break
			}
			log.Error().Err(streamErr).Msg("streaming subscriptions")
			hasErrorHappened = true
		}
	}

	if hasErrorHappened {
		return fmt.Errorf("some notifications failed")
	}

	return nil
}

func (e *WeatherJobExecutor) handleSubscription(
	ctx context.Context,
	sub *domain.Subscription,
) error {
	evt := &subscription.WeatherUpdatedEvent{
		Email:     sub.Email,
		City:      sub.City,
		Frequency: sub.Frequency,
		Token:     sub.Token,
	}
	return e.dispatcher.Dispatch(ctx, evt)
}

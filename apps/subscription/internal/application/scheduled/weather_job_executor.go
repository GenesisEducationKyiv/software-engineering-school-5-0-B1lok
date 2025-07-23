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

type GroupedSubscriptionReader interface {
	FindGroupedSubscriptions(
		ctx context.Context,
		frequency *domain.Frequency,
	) ([]*domain.GroupedSubscription, error)
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event event.Event) error
}

type WeatherJobExecutor struct {
	subscriptionRepo GroupedSubscriptionReader
	frequency        domain.Frequency
	dispatcher       EventDispatcher
}

func NewWeatherJobExecutor(
	subscriptionRepo GroupedSubscriptionReader,
	frequency domain.Frequency,
	dispatcher EventDispatcher,
) *WeatherJobExecutor {
	return &WeatherJobExecutor{
		subscriptionRepo: subscriptionRepo,
		frequency:        frequency,
		dispatcher:       dispatcher,
	}
}

func (e *WeatherJobExecutor) Execute(ctx context.Context) error {
	log.Info().Str("frequency", string(e.frequency)).Msg("Weather job started")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	hasErrorHappened := false
	groupedSubscriptions, err := e.subscriptionRepo.FindGroupedSubscriptions(ctx, &e.frequency)
	if err != nil {
		return err
	}

	for _, group := range groupedSubscriptions {
		for _, sub := range group.Subscriptions {
			select {
			case <-ctx.Done():
				log.Info().Err(ctx.Err()).Msg("context cancelled")
				return ctx.Err()
			default:
				err := e.dispatcher.Dispatch(ctx, &subscription.WeatherUpdatedEvent{
					Email:     sub.Email,
					City:      sub.City,
					Frequency: sub.Frequency,
					Token:     sub.Token,
				})
				if err != nil {
					log.Error().Err(err).Msg("failed to dispatch weather update")
					hasErrorHappened = true
				}
			}
		}
	}

	if hasErrorHappened {
		return fmt.Errorf("some notifications failed for frequency: %s", e.frequency)
	}

	log.Info().Str("frequency", string(e.frequency)).Msg("Weather job completed successfully")
	return nil
}

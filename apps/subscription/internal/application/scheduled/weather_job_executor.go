package scheduled

import (
	"context"
	"fmt"
	"log"
	"time"

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
	log.Printf("Weather job started for frequency: %s", e.frequency)

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
				log.Printf("context cancelled: %v", ctx.Err())
				return ctx.Err()
			default:
				err := e.dispatcher.Dispatch(ctx, &subscription.WeatherUpdatedEvent{
					Email:     sub.Email,
					City:      sub.City,
					Frequency: sub.Frequency,
					Token:     sub.Token,
				})
				if err != nil {
					log.Printf("failed to dispatch weather update: %v", err)
					hasErrorHappened = true
				}
			}
		}
	}

	if hasErrorHappened {
		return fmt.Errorf("some notifications failed for frequency: %s", e.frequency)
	}

	log.Printf("Weather job completed successfully for frequency: %s", e.frequency)
	return nil
}

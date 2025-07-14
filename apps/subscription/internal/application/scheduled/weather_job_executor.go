package scheduled

import (
	"context"
	"fmt"
	"log"
	"time"

	"subscription-service/internal/domain"
)

type GroupedSubscriptionReader interface {
	FindGroupedSubscriptions(
		ctx context.Context, frequency *domain.Frequency,
	) ([]*domain.GroupedSubscription, error)
}

type Notifier interface {
	NotifyWeatherUpdate(subscription *domain.Subscription) error
}

type WeatherJobExecutor struct {
	subscriptionRepo GroupedSubscriptionReader
	frequency        domain.Frequency
	notifier         Notifier
}

func NewWeatherJobExecutor(
	subscriptionRepo GroupedSubscriptionReader,
	frequency domain.Frequency,
	notifier Notifier,
) *WeatherJobExecutor {
	return &WeatherJobExecutor{
		subscriptionRepo: subscriptionRepo,
		frequency:        frequency,
		notifier:         notifier,
	}
}

func (e *WeatherJobExecutor) Execute(ctx context.Context) error {
	log.Printf("Weather job started for frequency: %s", e.frequency)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	errorHappened := false
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
				err := e.notifier.NotifyWeatherUpdate(sub)
				if err != nil {
					log.Printf("failed to notify weather update: %v", err)
					errorHappened = true
				}
			}
		}
	}

	if errorHappened {
		return fmt.Errorf("some notifications failed for frequency: %s", e.frequency)
	}

	log.Printf("Weather job completed successfully for frequency: %s", e.frequency)
	return nil
}

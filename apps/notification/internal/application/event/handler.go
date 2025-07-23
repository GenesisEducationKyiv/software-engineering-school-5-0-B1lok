package event

import (
	"context"
	"fmt"
)

type Name string

const UserSubscribedEventName Name = "user_subscribed"
const WeatherUpdatedEventName Name = "weather_updated"

type Handler interface {
	Handle(ctx context.Context, payload []byte) error
	GetName() Name
}

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

type Dispatcher struct {
	consumer Consumer
}

func NewDispatcher(consumer Consumer) *Dispatcher {
	return &Dispatcher{
		consumer: consumer,
	}
}

func (d *Dispatcher) Register(ctx context.Context, handler Handler) error {
	err := d.consumer.Consume(ctx, handler)
	if err != nil {
		return fmt.Errorf("failed to register handler for event %s: %w", handler.GetName(), err)
	}
	return nil
}

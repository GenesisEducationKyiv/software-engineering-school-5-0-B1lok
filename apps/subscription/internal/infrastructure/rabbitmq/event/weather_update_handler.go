package event

import (
	"context"
	"encoding/json"
	"fmt"

	"subscription-service/internal/application/event"
	"subscription-service/internal/application/event/subscription"
)

type Publisher interface {
	Publish(ctx context.Context, queue string, payload []byte) error
}

type WeatherUpdateHandler struct {
	host      string
	publisher Publisher
}

type weatherUpdatedPayload struct {
	Email          string `json:"email"`
	City           string `json:"city"`
	Frequency      string `json:"frequency"`
	UnsubscribeURL string `json:"unsubscribe_url"`
}

func NewWeatherUpdateHandler(host string, publisher Publisher) *WeatherUpdateHandler {
	return &WeatherUpdateHandler{
		host:      host,
		publisher: publisher,
	}
}

func (w *WeatherUpdateHandler) Handle(ctx context.Context, evt event.Event) error {
	e, ok := evt.(*subscription.WeatherUpdatedEvent)
	if !ok {
		return fmt.Errorf("invalid evt type: expected WeatherUpdatedEvent, got %T", evt)
	}
	payload, err := json.Marshal(w.toWeatherUpdatedPayload(e))
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := w.publisher.Publish(
		ctx,
		string(subscription.WeatherUpdatedEventName),
		payload,
	); err != nil {
		return fmt.Errorf("failed to publish weather update: %w", err)
	}
	return nil
}

func (w *WeatherUpdateHandler) CanHandle(name event.Name) bool {
	return name == subscription.WeatherUpdatedEventName
}

func (w *WeatherUpdateHandler) toWeatherUpdatedPayload(
	e *subscription.WeatherUpdatedEvent,
) weatherUpdatedPayload {
	return weatherUpdatedPayload{
		Email:          e.Email,
		City:           e.City,
		Frequency:      string(e.Frequency),
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", w.host, e.Token),
	}
}

package subscription

import (
	"subscription-service/internal/application/event"
	"subscription-service/internal/domain"
)

const WeatherUpdatedEventName event.Name = "weather_updated"

type WeatherUpdatedEvent struct {
	Email     string
	City      string
	Frequency domain.Frequency
	Token     string
}

func (w WeatherUpdatedEvent) EventName() event.Name {
	return WeatherUpdatedEventName
}

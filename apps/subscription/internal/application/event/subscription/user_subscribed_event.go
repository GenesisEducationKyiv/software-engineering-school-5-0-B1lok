package subscription

import (
	"subscription-service/internal/application/event"
	"subscription-service/internal/domain"
)

const UserSubscribedEventName event.Name = "user_subscribed"

type UserSubscribedEvent struct {
	ID        uint
	Email     string
	City      string
	Frequency domain.Frequency
	Token     string
}

func (u UserSubscribedEvent) EventName() event.Name {
	return UserSubscribedEventName
}

package command

import (
	"subscription-service/internal/domain"
)

type SubscribeCommand struct {
	Email     string
	City      string
	Frequency string
}

func (c *SubscribeCommand) ToSubscriptionLookup() *domain.SubscriptionLookup {
	return &domain.SubscriptionLookup{
		Email:     c.Email,
		City:      c.City,
		Frequency: domain.Frequency(c.Frequency),
	}
}

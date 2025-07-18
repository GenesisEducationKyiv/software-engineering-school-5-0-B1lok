package postgres

import (
	"time"

	"subscription-service/internal/domain"
)

type SubscriptionEntity struct {
	ID        uint `gorm:"primaryKey"`
	Email     string
	City      string
	Frequency domain.Frequency `gorm:"type:frequency_enum"`
	Token     string
	Confirmed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SubscriptionEntity) TableName() string {
	return "subscriptions"
}

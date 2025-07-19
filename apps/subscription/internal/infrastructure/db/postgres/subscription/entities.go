package subscription

import (
	"time"

	"subscription-service/internal/domain"
)

type Entity struct {
	ID        uint `gorm:"primaryKey"`
	Email     string
	City      string
	Frequency domain.Frequency `gorm:"type:frequency_enum"`
	Token     string
	Confirmed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Entity) TableName() string {
	return "subscriptions"
}

package outbox

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/datatypes"
)

type Status string
type EventType string

const (
	StatusPending Status = "pending"
	StatusFailed  Status = "failed"
	StatusSuccess Status = "success"

	EventUserSubscribed EventType = "user_subscribed"
)

type Message struct {
	ID          uint           `gorm:"primaryKey"`
	AggregateID uint           `gorm:"not null"`
	MessageID   uuid.UUID      `gorm:"type:uuid"`
	EventType   EventType      `gorm:"type:event_type_enum;not null"`
	Payload     datatypes.JSON `gorm:"type:jsonb;not null"`
	Status      Status         `gorm:"type:status_enum;default:'new'"`
	CreatedAt   time.Time      `gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt   time.Time      `gorm:"type:timestamp with time zone;default:now()"`
}

func (Message) TableName() string {
	return "outbox"
}

package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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
	ID          uint
	AggregateID uint
	MessageID   uuid.UUID
	EventType   EventType
	Payload     json.RawMessage
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

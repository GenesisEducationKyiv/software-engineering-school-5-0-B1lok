package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"subscription-service/internal/application/event"
	"subscription-service/internal/application/event/subscription"
	"subscription-service/internal/infrastructure/db/postgres/outbox"
)

type Saver interface {
	Save(ctx context.Context, outbox outbox.Message) error
}

type UserSubscribedHandler struct {
	host  string
	saver Saver
}

func NewUserSubscribedHandler(host string, saver Saver) *UserSubscribedHandler {
	return &UserSubscribedHandler{
		host:  host,
		saver: saver,
	}
}

type userSubscribedPayload struct {
	MessageId  uuid.UUID `json:"message_id"`
	Email      string    `json:"email"`
	City       string    `json:"city"`
	Frequency  string    `json:"frequency"`
	ConfirmUrl string    `json:"url"`
}

func (u *UserSubscribedHandler) Handle(ctx context.Context, evt event.Event) error {
	e, ok := evt.(*subscription.UserSubscribedEvent)
	if !ok {
		return fmt.Errorf("invalid evt type: expected UserSubscribedEvent, got %T", evt)
	}

	msg, err := u.toOutboxMessage(e)
	if err != nil {
		return fmt.Errorf("build outbox message: %w", err)
	}

	if err := u.saver.Save(ctx, msg); err != nil {
		return fmt.Errorf("failed to save outbox message: %w", err)
	}

	return nil
}

func (u *UserSubscribedHandler) CanHandle(name event.Name) bool {
	return name == subscription.UserSubscribedEventName
}

func (u *UserSubscribedHandler) toOutboxMessage(
	e *subscription.UserSubscribedEvent,
) (outbox.Message, error) {
	payload := userSubscribedPayload{
		MessageId:  uuid.New(),
		Email:      e.Email,
		City:       e.City,
		Frequency:  string(e.Frequency),
		ConfirmUrl: fmt.Sprintf("%s/api/confirm/%s", u.host, e.Token),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return outbox.Message{}, err
	}

	return outbox.Message{
		AggregateID: e.ID,
		EventType:   outbox.EventUserSubscribed,
		Payload:     data,
		Status:      outbox.StatusPending,
	}, nil
}

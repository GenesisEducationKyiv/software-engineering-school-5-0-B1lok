package event

import (
	"context"
	"encoding/json"
	"fmt"

	"notification/internal/application/event"
)

const templateName = "confirm.html"

type Sender interface {
	Send(templateName, to, subject string, data any) error
}

type UserSubscribedHandler struct {
	sender Sender
}

func NewUserSubscribedHandler(sender Sender) *UserSubscribedHandler {
	return &UserSubscribedHandler{
		sender: sender,
	}
}

type confirmationEmailTemplate struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	URL       string `json:"url"`
}

func (h *UserSubscribedHandler) Handle(ctx context.Context, payload []byte) error {
	var template confirmationEmailTemplate
	if err := json.Unmarshal(payload, &template); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := h.sender.Send(templateName, template.Email, "Confirm your subscription", template)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func (h *UserSubscribedHandler) GetName() event.Name {
	return event.UserSubscribedEventName
}

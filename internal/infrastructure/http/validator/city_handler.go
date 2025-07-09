package validator

import "context"

type Handler struct {
	client Client
	next   Client
}

func NewHandler(client Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) SetNext(next Client) {
	h.next = next
}

func (h *Handler) Validate(ctx context.Context, city string) (*string, error) {
	resp, err := h.client.Validate(ctx, city)
	if err != nil && h.next != nil {
		return h.next.Validate(ctx, city)
	}
	return resp, err
}

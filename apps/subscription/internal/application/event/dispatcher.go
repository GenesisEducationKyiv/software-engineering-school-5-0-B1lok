package event

import (
	"context"
	"fmt"
)

type Name string

type Event interface {
	EventName() Name
}

type Handler interface {
	Handle(ctx context.Context, event Event) error
	CanHandle(name Name) bool
}

type Dispatcher struct {
	handlers []Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make([]Handler, 0),
	}
}

func (d *Dispatcher) Register(handler Handler) {
	d.handlers = append(d.handlers, handler)
}

func (d *Dispatcher) Dispatch(ctx context.Context, event Event) error {
	for _, handler := range d.handlers {
		if handler.CanHandle(event.EventName()) {
			return handler.Handle(ctx, event)
		}
	}
	return fmt.Errorf("no handler found for event: %s", event.EventName())
}

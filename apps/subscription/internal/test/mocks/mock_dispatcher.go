package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"subscription-service/internal/application/event"
)

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(ctx context.Context, e event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

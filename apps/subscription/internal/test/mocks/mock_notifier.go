package mocks

import (
	"github.com/stretchr/testify/mock"

	"subscription-service/internal/domain"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) NotifyConfirmation(
	subscription *domain.Subscription,
) error {
	args := m.Called(subscription)
	return args.Error(0)
}

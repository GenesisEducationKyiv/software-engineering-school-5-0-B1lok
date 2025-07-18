package stubs

import "subscription-service/internal/domain"

type ConfirmationNotifierStub struct {
	NotifyConfirmationFn func(subscription *domain.Subscription) error
}

func NewConfirmationNotifierStub() *ConfirmationNotifierStub {
	return &ConfirmationNotifierStub{
		NotifyConfirmationFn: nil,
	}
}

func (s *ConfirmationNotifierStub) NotifyConfirmation(subscription *domain.Subscription) error {
	if s.NotifyConfirmationFn != nil {
		return s.NotifyConfirmationFn(subscription)
	}
	return nil
}

package mocks

import (
	"context"

	"subscription-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) ExistByLookup(ctx context.Context,
	lookup *domain.SubscriptionLookup,
) (bool, error) {
	args := m.Called(ctx, lookup)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionRepository) Create(ctx context.Context,
	subscription *domain.Subscription,
) (*domain.Subscription, error) {
	args := m.Called(ctx, subscription)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByToken(ctx context.Context,
	token string,
) (*domain.Subscription, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context,
	subscription *domain.Subscription,
) (*domain.Subscription, error) {
	args := m.Called(ctx, subscription)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

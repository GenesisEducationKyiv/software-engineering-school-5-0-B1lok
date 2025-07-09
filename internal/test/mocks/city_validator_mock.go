package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockCityValidator struct {
	mock.Mock
}

func (m *MockCityValidator) Validate(ctx context.Context, city string) (*string, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	validatedCity := args.Get(0).(string)
	return &validatedCity, args.Error(1)
}

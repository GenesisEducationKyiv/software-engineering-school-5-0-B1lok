package mocks

import "github.com/stretchr/testify/mock"

type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncActiveSubscriptions() {
	m.Called()
}

func (m *MockMetrics) DecActiveSubscriptions() {
	m.Called()
}

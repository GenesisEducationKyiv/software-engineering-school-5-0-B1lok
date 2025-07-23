//go:build unit
// +build unit

//nolint:gosec

package scheduled

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"subscription-service/internal/application/event"
	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	"subscription-service/internal/test/mocks"
	pkgErrors "subscription-service/pkg/errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEventDispatcher struct {
	mu            sync.Mutex
	callCount     int
	dispatched    []event.Event
	dispatchError error
}

func (m *mockEventDispatcher) Dispatch(ctx context.Context, e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.dispatched = append(m.dispatched, e)

	if m.dispatchError != nil {
		return m.dispatchError
	}
	return nil
}

func (m *mockEventDispatcher) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockEventDispatcher) GetDispatchedEvents() []event.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	events := make([]event.Event, len(m.dispatched))
	copy(events, m.dispatched)
	return events
}

func createTestSubscriptionsForCity(startID uint, city string, count int) []*domain.Subscription {
	subscriptions := make([]*domain.Subscription, count)
	for i := 0; i < count; i++ {
		subscriptions[i] = createTestSubscription(startID+uint(i), city)
	}
	return subscriptions
}

func createTestGroupedSubscriptionsDynamic(
	config map[string]int, startID uint,
) []*domain.GroupedSubscription {
	grouped := make([]*domain.GroupedSubscription, 0, len(config))
	currentID := startID

	for city, count := range config {
		subs := createTestSubscriptionsForCity(currentID, city, count)
		grouped = append(grouped, &domain.GroupedSubscription{
			City:          city,
			Subscriptions: subs,
		})
		currentID += uint(count)
	}

	return grouped
}

func generateRandomEmail() string {
	return fmt.Sprintf("user%s@example.com", uuid.New().String())
}

func generateRandomToken() string {
	return uuid.New().String()
}

func createTestSubscription(id uint, city string) *domain.Subscription {
	return &domain.Subscription{
		ID:        id,
		Email:     generateRandomEmail(),
		City:      city,
		Frequency: "hourly",
		Token:     generateRandomToken(),
		Confirmed: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func generateRandomCitySubCounts(cities []string, maxPerCity int) map[string]int {
	counts := make(map[string]int)
	for _, city := range cities {
		counts[city] = rand.Intn(maxPerCity) + 1
	}
	return counts
}

func calculateExpectedSubIDs(counts map[string]int) []uint {
	total := 0
	for _, c := range counts {
		total += c
	}
	ids := make([]uint, total)
	for i := range ids {
		ids[i] = uint(i + 1)
	}
	return ids
}

func TestWeatherJobExecutor_Execute_Success(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 100)
	groupedSubscriptions := createTestGroupedSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(groupedSubscriptions, nil).
		Once()

	mockDispatch := &mockEventDispatcher{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatch,
	)

	err := executor.Execute(ctx)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)

	expectedSubIDs := calculateExpectedSubIDs(citySubCounts)
	dispatchedEvents := mockDispatch.GetDispatchedEvents()

	assert.Equal(t, len(expectedSubIDs), mockDispatch.GetCallCount(),
		"mockDispatch should be called once per subscription")

	assert.Equal(t, len(expectedSubIDs), len(dispatchedEvents))
}

func TestWeatherJobExecutor_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	expectedError := pkgErrors.New(internalErrors.ErrInternal, "repository error")

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(nil, expectedError).
		Once()

	mockDispatcher := &mockEventDispatcher{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatcher,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockRepo.AssertExpectations(t)

	assert.Equal(t, 0, mockDispatcher.GetCallCount(),
		"dispatcherFunc should not be called when repository fails")
}

func TestWeatherJobExecutor_Execute_DispatcherError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 10)
	groupedSubscriptions := createTestGroupedSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(groupedSubscriptions, nil).
		Once()

	mockDispatcher := &mockEventDispatcher{
		dispatchError: pkgErrors.New(internalErrors.ErrInternal, "dispatcher error"),
	}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatcher,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when dispatcher fail")

	mockRepo.AssertExpectations(t)

	expectedSubIDs := calculateExpectedSubIDs(citySubCounts)
	assert.Equal(t, len(expectedSubIDs), mockDispatcher.GetCallCount(),
		"At least one dispatch should be attempted")
}

func TestWeatherJobExecutor_Execute_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 10)
	groupedSubscriptions := createTestGroupedSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(groupedSubscriptions, nil).
		Maybe()

	mockDispatcher := &mockEventDispatcher{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatcher,
	)

	time.Sleep(10 * time.Millisecond)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when context times out")
	assert.Equal(t, context.DeadlineExceeded, err)
}

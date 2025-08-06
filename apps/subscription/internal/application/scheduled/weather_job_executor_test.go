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
	pkgErrors "subscription-service/pkg/errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

type mockSubscriptionRepository struct {
	Subscriptions []*domain.Subscription
	ErrCh         chan error
	StartErr      error
}

func (m *mockSubscriptionRepository) StreamSubscriptions(
	ctx context.Context,
	_ *domain.Frequency,
) (<-chan domain.Subscription, <-chan error, error) {

	if m.StartErr != nil {
		return nil, nil, m.StartErr
	}

	subCh := make(chan domain.Subscription)
	errCh := make(chan error, 1)

	go func() {
		defer close(subCh)
		defer close(errCh)

		for _, sub := range m.Subscriptions {
			select {
			case <-ctx.Done():
				return
			case subCh <- *sub:
			}
		}

		if m.ErrCh != nil {
			for err := range m.ErrCh {
				errCh <- err
			}
		}
	}()

	return subCh, errCh, nil
}

func createTestSubscriptionsForCity(startID uint, city string, count int) []*domain.Subscription {
	subscriptions := make([]*domain.Subscription, count)
	for i := 0; i < count; i++ {
		subscriptions[i] = createTestSubscription(startID+uint(i), city)
	}
	return subscriptions
}

func createTestSubscriptionsDynamic(
	config map[string]int, startID uint,
) []*domain.Subscription {
	var subscriptions []*domain.Subscription
	currentID := startID

	for city, count := range config {
		subs := createTestSubscriptionsForCity(currentID, city, count)
		subscriptions = append(subscriptions, subs...)
		currentID += uint(count)
	}

	return subscriptions
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
	subscriptions := createTestSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := &mockSubscriptionRepository{}
	mockRepo.Subscriptions = subscriptions

	mockDispatch := &mockEventDispatcher{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatch,
	)

	err := executor.Execute(ctx)

	assert.NoError(t, err)

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

	mockRepo := &mockSubscriptionRepository{}
	mockRepo.StartErr = expectedError

	mockDispatcher := &mockEventDispatcher{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockDispatcher,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	assert.Equal(t, 0, mockDispatcher.GetCallCount(),
		"dispatcherFunc should not be called when repository fails")
}

func TestWeatherJobExecutor_Execute_DispatcherError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 10)
	subscriptions := createTestSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := &mockSubscriptionRepository{}
	mockRepo.Subscriptions = subscriptions

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
	subscriptions := createTestSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := &mockSubscriptionRepository{}
	mockRepo.Subscriptions = subscriptions

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

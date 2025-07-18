//go:build unit
// +build unit

//nolint:gosec

package scheduled

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
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

type mockNotifyFunc struct {
	mu                    sync.Mutex
	callCount             int
	notifiedSubscriptions []*domain.Subscription
	notifyError           error
}

func (m *mockNotifyFunc) NotifyWeatherUpdate(sub *domain.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.notifiedSubscriptions = append(m.notifiedSubscriptions, sub)

	if m.notifyError != nil {
		return m.notifyError
	}
	return nil
}

func (m *mockNotifyFunc) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockNotifyFunc) GetNotifiedSubscriptions() []*domain.Subscription {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := make([]*domain.Subscription, len(m.notifiedSubscriptions))
	copy(subs, m.notifiedSubscriptions)
	return subs
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

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockNotify,
	)

	err := executor.Execute(ctx)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)

	expectedSubIDs := calculateExpectedSubIDs(citySubCounts)
	notifiedSubs := mockNotify.GetNotifiedSubscriptions()

	assert.Equal(t, len(expectedSubIDs), mockNotify.GetCallCount(),
		"notifyFunc should be called once per subscription")

	assert.Equal(t, len(expectedSubIDs), len(notifiedSubs))

	actualSubIDs := make(map[uint]bool)
	for _, sub := range notifiedSubs {
		actualSubIDs[sub.ID] = true
	}
	for _, id := range expectedSubIDs {
		assert.True(t, actualSubIDs[id], "Expected subscription %d to be notified", id)
	}
}

func TestWeatherJobExecutor_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	expectedError := pkgErrors.New(internalErrors.ErrInternal, "repository error")

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(nil, expectedError).
		Once()

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockNotify,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockRepo.AssertExpectations(t)

	assert.Equal(t, 0, mockNotify.GetCallCount(),
		"notifyFunc should not be called when repository fails")
}

func TestWeatherJobExecutor_Execute_NotifyError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 10)
	groupedSubscriptions := createTestGroupedSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(groupedSubscriptions, nil).
		Once()

	mockNotify := &mockNotifyFunc{
		notifyError: pkgErrors.New(internalErrors.ErrInternal, "notification error"),
	}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockNotify,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when notifications fail")

	mockRepo.AssertExpectations(t)

	expectedSubIDs := calculateExpectedSubIDs(citySubCounts)
	assert.Equal(t, len(expectedSubIDs), mockNotify.GetCallCount(),
		"At least one notification should be attempted")
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

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor(
		mockRepo,
		frequency,
		mockNotify,
	)

	time.Sleep(10 * time.Millisecond)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when context times out")
	assert.Equal(t, context.DeadlineExceeded, err)
}

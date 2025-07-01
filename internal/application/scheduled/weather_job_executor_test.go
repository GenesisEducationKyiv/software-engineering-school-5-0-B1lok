//go:build unit
// +build unit

//nolint:gosec

package scheduled

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"sync"
	"testing"
	"time"

	internalErrors "weather-api/internal/errors"
	pkgErrors "weather-api/pkg/errors"

	"weather-api/internal/domain"
	"weather-api/internal/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWeatherFunc struct {
	mu           sync.Mutex
	callCount    int
	calledCities []string
	weatherError error
}

func (m *mockWeatherFunc) getWeather(city string) (*domain.WeatherHourly, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.calledCities = append(m.calledCities, city)

	if m.weatherError != nil {
		return nil, m.weatherError
	}
	return createTestWeatherData(city), nil
}

func (m *mockWeatherFunc) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockWeatherFunc) GetCalledCities() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cities := make([]string, len(m.calledCities))
	copy(cities, m.calledCities)
	return cities
}

type mockNotifyFunc struct {
	mu                    sync.Mutex
	callCount             int
	notifiedSubscriptions []*domain.Subscription
	weatherDataReceived   []*domain.WeatherHourly
	notifyError           error
}

func (m *mockNotifyFunc) notify(sub *domain.Subscription, weather *domain.WeatherHourly) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.notifiedSubscriptions = append(m.notifiedSubscriptions, sub)
	m.weatherDataReceived = append(m.weatherDataReceived, weather)

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

func (m *mockNotifyFunc) GetWeatherDataReceived() []*domain.WeatherHourly {
	m.mu.Lock()
	defer m.mu.Unlock()
	data := make([]*domain.WeatherHourly, len(m.weatherDataReceived))
	copy(data, m.weatherDataReceived)
	return data
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

func createTestWeatherData(location string) *domain.WeatherHourly {
	return &domain.WeatherHourly{
		Location:   location,
		Time:       time.Now().Format(time.RFC3339),
		TempC:      15.0 + rand.Float64()*20.0,
		WillItRain: rand.Intn(2) == 0,
		ChanceRain: rand.Intn(100),
		WillItSnow: rand.Intn(2) == 0,
		ChanceSnow: rand.Intn(100),
		Condition:  "Clear",
		Icon:       "clear",
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

	mockWeather := &mockWeatherFunc{}

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor[*domain.WeatherHourly](
		mockRepo,
		frequency,
		mockWeather.getWeather,
		mockNotify.notify,
	)

	err := executor.Execute(ctx)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)

	assert.Equal(t, len(cities), mockWeather.GetCallCount(),
		"getWeatherFunc should be called once per city")

	calledCities := mockWeather.GetCalledCities()
	for _, expectedCity := range cities {
		assert.Contains(t, calledCities, expectedCity,
			"Expected city %s to be called", expectedCity)
	}

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

	weatherDataReceived := mockNotify.GetWeatherDataReceived()
	assert.Equal(t, len(expectedSubIDs), len(weatherDataReceived))
}

func TestWeatherJobExecutor_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	expectedError := pkgErrors.New(internalErrors.ErrInternal, "repository error")

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(nil, expectedError).
		Once()

	mockWeather := &mockWeatherFunc{}

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor[*domain.WeatherHourly](
		mockRepo,
		frequency,
		mockWeather.getWeather,
		mockNotify.notify,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	mockRepo.AssertExpectations(t)

	assert.Equal(t, 0, mockWeather.GetCallCount(),
		"getWeatherFunc should not be called when repository fails")

	assert.Equal(t, 0, mockNotify.GetCallCount(),
		"notifyFunc should not be called when repository fails")
}

func TestWeatherJobExecutor_Execute_WeatherAPIError(t *testing.T) {
	ctx := context.Background()
	frequency := domain.Frequency("hourly")
	cities := []string{"New York", "London", "Tokyo", "Paris", "Berlin"}
	citySubCounts := generateRandomCitySubCounts(cities, 1)
	groupedSubscriptions := createTestGroupedSubscriptionsDynamic(citySubCounts, 1)

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockRepo.On("FindGroupedSubscriptions", mock.Anything, &frequency).
		Return(groupedSubscriptions, nil).
		Once()

	mockWeather := &mockWeatherFunc{
		weatherError: pkgErrors.New(internalErrors.ErrInternal, "weather API error"),
	}

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor[*domain.WeatherHourly](
		mockRepo,
		frequency,
		mockWeather.getWeather,
		mockNotify.notify,
	)

	err := executor.Execute(ctx)

	assert.NoError(t, err, "Weather API errors should be logged but not stop execution")

	mockRepo.AssertExpectations(t)

	assert.Equal(t, len(cities), mockWeather.GetCallCount(),
		"getWeatherFunc should be called for each city despite errors")

	assert.Equal(t, 0, mockNotify.GetCallCount(),
		"No notifications should be sent when weather API fails")
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

	mockWeather := &mockWeatherFunc{}

	mockNotify := &mockNotifyFunc{
		notifyError: pkgErrors.New(internalErrors.ErrInternal, "notification error"),
	}

	executor := NewWeatherJobExecutor[*domain.WeatherHourly](
		mockRepo,
		frequency,
		mockWeather.getWeather,
		mockNotify.notify,
	)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when notifications fail")

	mockRepo.AssertExpectations(t)

	assert.Equal(t, len(cities), mockWeather.GetCallCount(),
		"getWeatherFunc should be called for each city")

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

	mockWeather := &mockWeatherFunc{}

	mockNotify := &mockNotifyFunc{}

	executor := NewWeatherJobExecutor[*domain.WeatherHourly](
		mockRepo,
		frequency,
		mockWeather.getWeather,
		mockNotify.notify,
	)

	time.Sleep(10 * time.Millisecond)

	err := executor.Execute(ctx)

	assert.Error(t, err, "Should return error when context times out")
	assert.Equal(t, context.DeadlineExceeded, err)
}

//go:build unit
// +build unit

package subscription

import (
	"context"
	"subscription-service/internal/application/command"
	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	"subscription-service/internal/test/mocks"
	"subscription-service/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validToken    = "valid-token-123"
	ValidatedCity = "Berlin"
)

func TestSubscriptionService_Subscribe_Success(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "daily",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	lookup := &domain.SubscriptionLookup{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: domain.Frequency("daily"),
	}

	subscription := &domain.Subscription{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: domain.Frequency("daily"),
	}
	mockRepo.On("ExistByLookup", ctx, lookup).Return(false, nil)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(subscription, nil)

	mockNotifier.On("NotifyConfirmation", mock.AnythingOfType("*domain.Subscription")).Return(nil)

	err := service.Subscribe(ctx, cmd)

	assert.NoError(t, err)
	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestSubscriptionService_Subscribe_InvalidCity(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "invalidcity",
		Frequency: "daily",
	}

	validationErr := errors.New(internalErrors.ErrNotFound, "City not found")
	mockValidator.On("Validate", ctx, "invalidcity").Return(nil, validationErr)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, validationErr, err)
	mockValidator.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "ExistByLookup")
	mockRepo.AssertNotCalled(t, "Create")
	mockNotifier.AssertNotCalled(t, "NotifyConfirmation")
}

func TestSubscriptionService_Subscribe_AlreadyExists(t *testing.T) {

	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "daily",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	lookup := &domain.SubscriptionLookup{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: "daily",
	}
	mockRepo.On("ExistByLookup", ctx, lookup).Return(true, nil)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Email already subscribed")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrConflict, apiErr.Base)

	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockNotifier.AssertNotCalled(t, "NotifyConfirmation")
}

func TestSubscriptionService_Subscribe_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "daily",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	lookup := &domain.SubscriptionLookup{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: "daily",
	}
	repoErr := errors.New(internalErrors.ErrInternal, "failed to check if email exists")
	mockRepo.On("ExistByLookup", ctx, lookup).Return(false, repoErr)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if email exists")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInternal, apiErr.Base)

	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockNotifier.AssertNotCalled(t, "NotifyConfirmation")
}

func TestSubscriptionService_Subscribe_InvalidFrequency(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "invalid",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	validatedCmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: "invalid",
	}

	mockRepo.On("ExistByLookup", ctx, validatedCmd.ToSubscriptionLookup()).Return(false, nil)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid frequency value")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInvalidInput, apiErr.Base)

	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockNotifier.AssertNotCalled(t, "NotifyConfirmation")
}

func TestSubscriptionService_Subscribe_CreateError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "daily",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	lookup := &domain.SubscriptionLookup{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: "daily",
	}
	mockRepo.On("ExistByLookup", ctx, lookup).Return(false, nil)

	createErr := errors.New(internalErrors.ErrInternal, "Database error")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil, createErr)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Database error")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInternal, apiErr.Base)

	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertNotCalled(t, "NotifyConfirmation")
}

func TestSubscriptionService_Subscribe_EmailError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	cmd := &command.SubscribeCommand{
		Email:     "test@example.com",
		City:      "berlin",
		Frequency: "daily",
	}

	mockValidator.On("Validate", ctx, "berlin").Return(ValidatedCity, nil)

	lookup := &domain.SubscriptionLookup{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: "daily",
	}

	mockRepo.On("ExistByLookup", ctx, lookup).Return(false, nil)
	subscription := &domain.Subscription{
		Email:     "test@example.com",
		City:      ValidatedCity,
		Frequency: domain.Frequency("daily"),
	}
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(subscription, nil)

	emailErr := errors.New(internalErrors.ErrInternal, "Failed to send email")
	mockNotifier.On(
		"NotifyConfirmation", mock.AnythingOfType("*domain.Subscription"),
	).Return(emailErr)

	err := service.Subscribe(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, emailErr, err)

	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestSubscriptionService_Confirm_Success(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	subscription := &domain.Subscription{
		ID:        1,
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: domain.Frequency("daily"),
		Token:     token,
		Confirmed: false,
	}

	mockRepo.On("FindByToken", ctx, token).Return(subscription, nil)

	updatedSubscription := &domain.Subscription{
		ID:        1,
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: domain.Frequency("daily"),
		Token:     token,
		Confirmed: true,
	}
	mockRepo.On(
		"Update", ctx, mock.AnythingOfType("*domain.Subscription"),
	).Return(updatedSubscription, nil)

	err := service.Confirm(ctx, token)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_Confirm_TokenNotFound(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := "invalid-token"

	mockRepo.On("FindByToken", ctx, token).Return(nil, nil)

	err := service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token not found")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrNotFound, apiErr.Base)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestSubscriptionService_Confirm_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	repoErr := errors.New(internalErrors.ErrInternal, "Database error")
	mockRepo.On("FindByToken", ctx, token).Return(nil, repoErr)

	err := service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Database error")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInternal, apiErr.Base)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestSubscriptionService_Confirm_UpdateError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	subscription := &domain.Subscription{
		ID:        1,
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: domain.Frequency("daily"),
		Token:     token,
		Confirmed: false,
	}

	mockRepo.On("FindByToken", ctx, token).Return(subscription, nil)

	updateErr := errors.New(internalErrors.ErrInternal, "Update failed")
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil, updateErr)

	err := service.Confirm(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Update failed")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInternal, apiErr.Base)

	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_Unsubscribe_Success(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	subscription := &domain.Subscription{
		ID:        1,
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: domain.Frequency("daily"),
		Token:     token,
		Confirmed: true,
	}

	mockRepo.On("FindByToken", ctx, token).Return(subscription, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	err := service.Unsubscribe(ctx, token)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_Unsubscribe_TokenNotFound(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := "invalid-token"

	mockRepo.On("FindByToken", ctx, token).Return(nil, nil)

	err := service.Unsubscribe(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token not found")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrNotFound, apiErr.Base)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestSubscriptionService_Unsubscribe_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	repoErr := errors.New(internalErrors.ErrInternal, "Database error")
	mockRepo.On("FindByToken", ctx, token).Return(nil, repoErr)

	err := service.Unsubscribe(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Database error")
	apiErr, ok := errors.IsApiError(err)
	assert.True(t, ok)
	assert.Equal(t, internalErrors.ErrInternal, apiErr.Base)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestSubscriptionService_Unsubscribe_DeleteError(t *testing.T) {
	mockRepo := new(mocks.MockSubscriptionRepository)
	mockValidator := new(mocks.MockCityValidator)
	mockNotifier := new(mocks.MockNotifier)

	service := NewService(mockRepo, mockValidator, mockNotifier)

	ctx := context.Background()
	token := validToken

	subscription := &domain.Subscription{
		ID:        1,
		Email:     "test@example.com",
		City:      "Berlin",
		Frequency: domain.Frequency("daily"),
		Token:     token,
		Confirmed: true,
	}

	mockRepo.On("FindByToken", ctx, token).Return(subscription, nil)

	deleteErr := errors.New(internalErrors.ErrInternal, "Delete failed")
	mockRepo.On("Delete", ctx, uint(1)).Return(deleteErr)

	err := service.Unsubscribe(ctx, token)

	assert.Error(t, err)
	assert.Equal(t, deleteErr, err)

	mockRepo.AssertExpectations(t)
}

package postgres

import (
	"weather-api/internal/domain"
)

func toEntity(subscription *domain.Subscription) *SubscriptionEntity {
	if subscription == nil {
		return nil
	}

	return &SubscriptionEntity{
		ID:        subscription.ID,
		Email:     subscription.Email,
		City:      subscription.City,
		Frequency: subscription.Frequency,
		Token:     subscription.Token,
		Confirmed: subscription.Confirmed,
		CreatedAt: subscription.CreatedAt,
		UpdatedAt: subscription.UpdatedAt,
	}
}

func toDomain(entity *SubscriptionEntity) (*domain.Subscription, error) {
	if entity == nil {
		return nil, nil
	}

	subscription := &domain.Subscription{
		ID:        entity.ID,
		Email:     entity.Email,
		City:      entity.City,
		Frequency: entity.Frequency,
		Token:     entity.Token,
		Confirmed: entity.Confirmed,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}

	return subscription, nil
}

func toDomainList(entities []SubscriptionEntity) ([]*domain.Subscription, error) {
	if entities == nil {
		return nil, nil
	}

	result := make([]*domain.Subscription, 0, len(entities))
	for _, entity := range entities {
		sub, err := toDomain(&entity)
		if err != nil {
			return nil, err
		}
		result = append(result, sub)
	}

	return result, nil
}

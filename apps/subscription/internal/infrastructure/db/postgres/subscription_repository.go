package postgres

import (
	"context"
	"errors"

	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
	"subscription-service/pkg/middleware"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(
	ctx context.Context, subscription *domain.Subscription,
) (*domain.Subscription, error) {
	entity := toEntity(subscription)
	db := r.getDB(ctx)

	if err := db.Create(entity).Error; err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to create subscription",
		)
	}

	saved, err := toDomain(entity)
	if err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to map subscription entity",
		)
	}
	return saved, nil
}

func (r *SubscriptionRepository) ExistByLookup(
	ctx context.Context, lookup *domain.SubscriptionLookup,
) (bool, error) {
	var count int64
	db := r.getDB(ctx)
	err := db.
		Model(&SubscriptionEntity{}).
		Where("email = ? AND city = ? AND frequency = ?", lookup.Email, lookup.City, lookup.Frequency).
		Count(&count).Error
	if err != nil {
		return false, pkgErrors.New(
			internalErrors.ErrInternal, "failed to check if email exists",
		)
	}
	return count > 0, nil
}

func (r *SubscriptionRepository) Update(
	ctx context.Context, subscription *domain.Subscription,
) (*domain.Subscription, error) {
	entity := toEntity(subscription)
	db := r.getDB(ctx)
	result := db.Save(entity)
	if result.Error != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to update subscription",
		)
	}

	if result.RowsAffected == 0 {
		return nil, pkgErrors.New(internalErrors.ErrNotFound, "subscription not found")
	}

	return toDomain(entity)
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uint) error {
	db := r.getDB(ctx)
	result := db.Delete(&SubscriptionEntity{}, id)
	if result.Error != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to delete subscription",
		)
	}

	if result.RowsAffected == 0 {
		return pkgErrors.New(internalErrors.ErrNotFound, "subscription not found")
	}

	return nil
}

func (r *SubscriptionRepository) FindByToken(
	ctx context.Context, token string,
) (*domain.Subscription, error) {
	var entity SubscriptionEntity
	db := r.getDB(ctx)
	result := db.Where("token = ?", token).First(&entity)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to find subscription by token",
		)
	}

	return toDomain(&entity)
}

func (r *SubscriptionRepository) FindGroupedSubscriptions(
	ctx context.Context, frequency *domain.Frequency,
) ([]*domain.GroupedSubscription, error) {
	var subscriptions []SubscriptionEntity
	db := r.getDB(ctx)

	err := db.
		Where("confirmed = ? AND frequency = ?", true, frequency).
		Find(&subscriptions).Error
	if err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to find confirmed subscriptions",
		)
	}

	subscriptionMap := make(map[string][]SubscriptionEntity)
	for _, sub := range subscriptions {
		subscriptionMap[sub.City] = append(subscriptionMap[sub.City], sub)
	}

	grouped := make([]*domain.GroupedSubscription, 0, len(subscriptionMap))
	for city, subscriptions := range subscriptionMap {
		domainSubs, err := toDomainList(subscriptions)
		if err != nil {
			return nil, err
		}
		grouped = append(grouped, &domain.GroupedSubscription{
			City:          city,
			Subscriptions: domainSubs,
		})
	}

	return grouped, nil
}

func (r *SubscriptionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := middleware.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

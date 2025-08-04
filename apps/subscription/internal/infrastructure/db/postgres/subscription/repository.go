package subscription

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
	"subscription-service/pkg/middleware"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(
	ctx context.Context, subscription *domain.Subscription,
) (*domain.Subscription, error) {
	db := r.getDB(ctx)

	query := `
        INSERT INTO subscriptions (email, city, frequency, token, confirmed, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at`

	now := time.Now()
	var id uint
	var createdAt, updatedAt time.Time

	err := db.QueryRow(
		ctx, query,
		subscription.Email,
		subscription.City,
		subscription.Frequency,
		subscription.Token,
		subscription.Confirmed,
		now,
		now,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to create subscription",
		)
	}
	subscription.ID = id
	subscription.CreatedAt = createdAt
	subscription.UpdatedAt = updatedAt

	return subscription, nil
}

func (r *Repository) ExistByLookup(
	ctx context.Context, lookup *domain.SubscriptionLookup,
) (bool, error) {
	const query = `
        SELECT EXISTS (
            SELECT 1
            FROM subscriptions
            WHERE email = $1 AND city = $2 AND frequency = $3
        )`

	db := r.getDB(ctx)

	var exists bool
	err := db.QueryRow(ctx, query, lookup.Email, lookup.City, lookup.Frequency).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking subscription existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) Update(
	ctx context.Context, subscription *domain.Subscription,
) (*domain.Subscription, error) {
	db := r.getDB(ctx)
	query := `
        UPDATE subscriptions 
        SET email = $2, city = $3, frequency = $4, token = $5, confirmed = $6, updated_at = $7
        WHERE id = $1
        RETURNING updated_at`

	now := time.Now()
	var updatedAt time.Time

	err := db.QueryRow(
		ctx, query,
		subscription.ID,
		subscription.Email,
		subscription.City,
		subscription.Frequency,
		subscription.Token,
		subscription.Confirmed,
		now,
	).Scan(&updatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgErrors.New(internalErrors.ErrNotFound, "subscription not found")
		}
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to update subscription",
		)
	}

	subscription.UpdatedAt = updatedAt

	return subscription, nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	db := r.getDB(ctx)
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := db.Exec(ctx, query, id)
	if err != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to delete subscription",
		)
	}

	if result.RowsAffected() == 0 {
		return pkgErrors.New(internalErrors.ErrNotFound, "subscription not found")
	}

	return nil
}

func (r *Repository) FindByToken(
	ctx context.Context, token string,
) (*domain.Subscription, error) {
	db := r.getDB(ctx)
	query := `
        SELECT id, email, city, frequency, token, confirmed, created_at, updated_at
        FROM subscriptions 
        WHERE token = $1`

	var subscription domain.Subscription
	err := db.QueryRow(ctx, query, token).Scan(
		&subscription.ID,
		&subscription.Email,
		&subscription.City,
		&subscription.Frequency,
		&subscription.Token,
		&subscription.Confirmed,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to find subscription by token",
		)
	}

	return &subscription, nil
}

func (r *Repository) FindGroupedSubscriptions(
	ctx context.Context, frequency *domain.Frequency,
) ([]*domain.GroupedSubscription, error) {
	var subscriptions []Entity
	db := r.getDB(ctx)

	err := db.
		Where("confirmed = ? AND frequency = ?", true, frequency).
		Find(&subscriptions).Error
	if err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to find confirmed subscriptions",
		)
	}

	subscriptionMap := make(map[string][]Entity)
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

func (r *Repository) getDB(ctx context.Context) *pgxpool.Pool {
	if tx, ok := middleware.GetTx(ctx); ok {
		return tx
	}
	return r.pool
}

package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"subscription-service/internal/domain"
	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
	"subscription-service/pkg/middleware"
)

const (
	channelSize = 1000
	fetchSize   = 1000
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

func (r *Repository) StreamSubscriptions(
	ctx context.Context, frequency *domain.Frequency,
) (<-chan domain.Subscription, <-chan error, error) {
	subCh := make(chan domain.Subscription, channelSize)
	errCh := make(chan error, 1)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}

	cursorName := fmt.Sprintf("subscriptions_cursor_%d", time.Now().UnixNano())

	if err := declareCursor(ctx, tx, cursorName, frequency); err != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("declare cursor: %w", err)
	}

	go func() {
		defer close(subCh)
		defer close(errCh)
		defer rollbackSafely(tx, ctx)

		for {
			count, err := fetchNext(ctx, tx, cursorName, subCh)
			if err != nil {
				errCh <- err
				return
			}
			if count == 0 {
				break
			}
		}

		if err := closeCursor(ctx, tx, cursorName); err != nil {
			errCh <- err
			return
		}
		if err := tx.Commit(ctx); err != nil {
			errCh <- fmt.Errorf("failed to commit transaction: %w", err)
			return
		}
	}()

	return subCh, errCh, nil
}

func declareCursor(
	ctx context.Context,
	tx pgx.Tx,
	cursorName string,
	freq *domain.Frequency,
) error {
	query := fmt.Sprintf(`
		DECLARE %s CURSOR FOR
		SELECT id, email, city, frequency, token, confirmed, created_at, updated_at
		FROM subscriptions
		WHERE confirmed = true AND frequency = $1
		ORDER BY id`, cursorName)
	_, err := tx.Exec(ctx, query, freq)
	return err
}

func closeCursor(ctx context.Context, tx pgx.Tx, cursorName string) error {
	query := fmt.Sprintf("CLOSE %s", cursorName)
	_, err := tx.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("close cursor: %w", err)
	}
	return nil
}

func fetchNext(
	ctx context.Context,
	tx pgx.Tx,
	cursorName string,
	out chan<- domain.Subscription,
) (int, error) {
	query := fmt.Sprintf("FETCH %d FROM %s", fetchSize, cursorName)
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("fetch from cursor: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return 0, fmt.Errorf("scan subscription: %w", err)
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case out <- sub:
			count++
		}
	}
	return count, nil
}

func scanSubscription(row pgx.Row) (domain.Subscription, error) {
	var s domain.Subscription
	err := row.Scan(
		&s.ID,
		&s.Email,
		&s.City,
		&s.Frequency,
		&s.Token,
		&s.Confirmed,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	return s, err
}

func rollbackSafely(tx pgx.Tx, ctx context.Context) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Ctx(ctx).Error().Err(err).Msg("rollback failed")
	}
}

func (r *Repository) getDB(ctx context.Context) *pgxpool.Pool {
	if tx, ok := middleware.GetTx(ctx); ok {
		return tx
	}
	return r.pool
}

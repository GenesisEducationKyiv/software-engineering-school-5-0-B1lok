package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
	"subscription-service/pkg/middleware"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, outbox Message) error {
	db := r.getDB(ctx)

	query := `
        INSERT INTO outbox (
            aggregate_id, event_type, payload, status, message_id, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7
        )
        RETURNING id, created_at, updated_at`

	now := time.Now()
	_, err := db.Exec(
		ctx, query,
		outbox.AggregateID,
		outbox.EventType,
		outbox.Payload,
		outbox.Status,
		outbox.MessageID,
		now,
		now,
	)

	if err != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to create outbox message",
		)
	}

	return nil
}

func (r *Repository) GetPendingMessages(ctx context.Context, limit int) ([]Message, error) {
	db := r.getDB(ctx)

	const query = `
		SELECT id, aggregate_id, message_id, event_type, payload, status, created_at, updated_at
		FROM outbox
		WHERE status = $1 OR status = $2
		ORDER BY created_at
		LIMIT $3
		FOR UPDATE SKIP LOCKED
	`

	rows, err := db.Query(ctx, query, StatusPending, StatusFailed, limit)
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInternal, "failed to execute query: "+err.Error())
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(
			&m.ID,
			&m.AggregateID,
			&m.MessageID,
			&m.EventType,
			&m.Payload,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, pkgErrors.New(internalErrors.ErrInternal, "failed to scan row: "+err.Error())
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInternal, "row iteration error: "+err.Error())
	}

	return messages, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, messageID uint, status Status) error {
	db := r.getDB(ctx)

	query := `
        UPDATE outbox 
        SET status = $1, updated_at = $2
        WHERE id = $3`

	now := time.Now()
	tag, err := db.Exec(ctx, query, status, now, messageID)
	if err != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to update message status",
		)
	}

	if tag.RowsAffected() == 0 {
		return pkgErrors.New(
			internalErrors.ErrInternal, "message not found or not updated",
		)
	}

	return nil
}

func (r *Repository) getDB(ctx context.Context) *pgxpool.Pool {
	if tx, ok := middleware.GetTx(ctx); ok {
		return tx
	}
	return r.pool
}

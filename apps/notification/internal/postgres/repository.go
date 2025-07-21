package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"notification/pkg"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r Repository) SaveMessageId(ctx context.Context, messageId string) error {
	tx, ok := pkg.GetTx(ctx)
	if !ok {
		return errors.New("no transaction found in context")
	}

	ct, err := tx.Exec(ctx, `
		INSERT INTO notification_idempotence (message_id)
		VALUES ($1)
		ON CONFLICT DO NOTHING
	`, messageId)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("message already processed")
	}

	return nil
}

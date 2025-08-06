package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"

	"subscription-service/internal/config"
)

type DB struct {
	Pool *pgxpool.Pool
}

func ConnectDB(ctx context.Context, cfg config.DBConfig) (*DB, error) {
	escapedPassword := url.QueryEscape(cfg.Password)
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, escapedPassword, cfg.Host, cfg.Port, cfg.Name,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

package postgres

import (
	"fmt"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"

	"notification/internal/config"
)

func RunMigrations(cfg config.DBConfig) {
	escapedPassword := url.QueryEscape(cfg.Password)
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, escapedPassword, cfg.Host, cfg.Port, cfg.Name)

	m, err := migrate.New("file://migrations", connectionString)
	if err != nil {
		log.Error().Err(err).Msg("Migration initialization failed")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error().Err(err).Msg("Migration failed")
	}
}

func RunMigrationsWithPath(cfg config.DBConfig, migrationPath string) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	m, err := migrate.New(migrationPath, connectionString)
	if err != nil {
		log.Error().Err(err).Msg("Migration initialization failed")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error().Err(err).Msg("Migration failed")
	}
}

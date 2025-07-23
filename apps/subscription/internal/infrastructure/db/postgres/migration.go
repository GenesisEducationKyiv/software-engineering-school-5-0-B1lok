package postgres

import (
	"fmt"
	"net/url"

	"github.com/rs/zerolog/log"

	"subscription-service/internal/config"

	"github.com/golang-migrate/migrate/v4"
)

func RunMigrations(cfg config.DBConfig) {
	escapedPassword := url.QueryEscape(cfg.Password)
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, escapedPassword, cfg.Host, cfg.Port, cfg.Name)

	m, err := migrate.New("file://migrations", connectionString)
	if err != nil {
		log.Error().Err(err).Msgf("Could not connect to migrations")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error().Err(err).Msgf("Could not run migrations")
	}
}

func RunMigrationsWithPath(cfg config.DBConfig, migrationPath string) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	m, err := migrate.New(migrationPath, connectionString)
	if err != nil {
		log.Error().Err(err).Msgf("Could not connect to migrations")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error().Err(err).Msgf("Could not run migrations")
	}
}

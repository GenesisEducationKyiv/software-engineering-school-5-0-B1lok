package postgres

import (
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"

	"notification/internal/config"
)

func RunMigrations(cfg config.DBConfig) {
	escapedPassword := url.QueryEscape(cfg.Password)
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, escapedPassword, cfg.Host, cfg.Port, cfg.Name)

	m, err := migrate.New("file://migrations", connectionString)
	if err != nil {
		log.Printf("Migration initialization failed: %v", err)
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Migration failed: %v", err)
	}
}

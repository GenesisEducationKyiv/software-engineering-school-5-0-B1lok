package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Container testcontainers.Container
	URI       string
	DB        *pgxpool.Pool
}

func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL(
			"5432/tcp", "postgres", func(host string, port nat.Port,
			) string {
				return fmt.Sprintf(
					"host=%s port=%s user=test password=test dbname=testdb sslmode=disable",
					host, port.Port())
			}).WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf(
		"host=%s user=test password=test dbname=testdb port=%s sslmode=disable",
		host, port.Port())

	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with pgxpool: %w", err)
	}

	return &PostgresContainer{Container: container, URI: uri, DB: pool}, nil
}

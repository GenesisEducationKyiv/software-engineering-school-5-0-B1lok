package containers

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresContainer struct {
	Container testcontainers.Container
	URI       string
	DB        *gorm.DB
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
		WaitingFor: wait.ForListeningPort("5432/tcp"),
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

	uri := fmt.Sprintf("host=%s user=test password=test dbname=testdb port=%s sslmode=disable", host, port.Port())

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{Container: container, URI: uri, DB: db}, nil
}

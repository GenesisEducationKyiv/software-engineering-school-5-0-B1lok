//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"log"
	"notification/internal/application/event"
	"notification/internal/config"
	infevent "notification/internal/infrastructure/event"
	"notification/internal/interface/rabbitmq"
	"notification/pkg"

	appPostgres "notification/internal/postgres"
	"notification/internal/test/containers"
	"notification/internal/test/stub"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type IdempotentConsumerTestSuite struct {
	suite.Suite
	Rabbit   *containers.RabbitMQContainer
	Postgres *containers.PostgresContainer
	Pool     *pgxpool.Pool
}

type confirmationEmailTemplate struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	URL       string `json:"url"`
}

func (suite *IdempotentConsumerTestSuite) SetupSuite() {
	ctx := context.Background()
	rabbit, err := containers.SetupRabbitMQContainer(ctx)
	suite.Require().NoError(err)
	suite.Rabbit = rabbit
	postgres, err := containers.SetupPostgresContainer(ctx)
	suite.Require().NoError(err)
	suite.Postgres = postgres

	host, err := postgres.Container.Host(ctx)
	suite.Require().NoError(err)

	mappedPort, err := postgres.Container.MappedPort(ctx, "5432")
	suite.Require().NoError(err)

	cfg := config.DBConfig{
		User:     "test",
		Password: "test",
		Host:     host,
		Port:     mappedPort.Port(),
		Name:     "testdb",
	}

	db, err := appPostgres.ConnectDB(ctx, cfg)
	suite.Require().NoError(err)
	suite.Pool = db.Pool
	appPostgres.RunMigrationsWithPath(cfg, getMigrationPath())
}

func (suite *IdempotentConsumerTestSuite) TearDownSuite() {
	if suite.Rabbit != nil {
		ctx := context.Background()
		suite.Rabbit.Cleanup(ctx)
	}
}

func (suite *IdempotentConsumerTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	suite.Rabbit.Channel.QueueDelete(
		string(event.UserSubscribedEventName),
		false,
		false,
		false,
	)
	_, err := suite.Pool.Exec(ctx, "DELETE FROM notification_idempotence")
	suite.Require().NoError(err)
}

func (suite *IdempotentConsumerTestSuite) TestUserSubscribed_Idempotency() {
	ctx := context.Background()

	senderCalled := 0
	mockSender := &stub.Sender{
		SendFunc: func(templateName, to, subject string, data any) error {
			senderCalled++
			return nil
		},
	}
	handler := infevent.NewUserSubscribedHandler(mockSender)

	repo := appPostgres.NewRepository(suite.Pool)
	txManager := pkg.NewTxManager(suite.Pool)
	consumer := rabbitmq.NewIdempotentConsumer(suite.Rabbit.Channel, txManager, repo)

	err := suite.Rabbit.DeclareQueue(string(event.UserSubscribedEventName))
	suite.Require().NoError(err)

	err = consumer.Consume(ctx, handler)
	suite.Require().NoError(err)

	testPayload := confirmationEmailTemplate{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
		URL:       "https://unsubscribe.example.com",
	}

	payloadBytes, err := json.Marshal(testPayload)
	suite.Require().NoError(err)

	headers := map[string]interface{}{"message_id": uuid.New().String()}
	for i := 0; i < 5; i++ {
		err := suite.Rabbit.PublishMessageWithHeaders(
			ctx,
			string(event.UserSubscribedEventName),
			payloadBytes,
			headers,
		)
		suite.Require().NoError(err)
	}

	suite.Eventually(func() bool {
		return senderCalled == 1
	}, 5*time.Second, 100*time.Millisecond, "Handler should be called once due to idempotency")

	suite.Equal(1, senderCalled, "Sender should be called only once with same message_id")
}

func (suite *IdempotentConsumerTestSuite) TestUserSubscribed_FailureDoesNotStoreMessageID() {
	ctx := context.Background()

	mockSender := &stub.Sender{
		SendFunc: func(templateName, to, subject string, data any) error {
			return fmt.Errorf("simulated failure")
		},
	}
	handler := infevent.NewUserSubscribedHandler(mockSender)

	repo := appPostgres.NewRepository(suite.Pool)
	txManager := pkg.NewTxManager(suite.Pool)
	consumer := rabbitmq.NewIdempotentConsumer(suite.Rabbit.Channel, txManager, repo)

	err := suite.Rabbit.DeclareQueue(string(event.UserSubscribedEventName))
	suite.Require().NoError(err)

	err = consumer.Consume(ctx, handler)
	suite.Require().NoError(err)

	testPayload := confirmationEmailTemplate{
		Email:     "fail@example.com",
		City:      "Lviv",
		Frequency: "daily",
		URL:       "https://unsubscribe.example.com",
	}
	payloadBytes, err := json.Marshal(testPayload)
	suite.Require().NoError(err)

	messageID := uuid.New().String()
	headers := map[string]interface{}{"message_id": messageID}

	err = suite.Rabbit.PublishMessageWithHeaders(
		ctx,
		string(event.UserSubscribedEventName),
		payloadBytes,
		headers,
	)
	suite.Require().NoError(err)

	var exists bool
	err = suite.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM notification_idempotence WHERE message_id = $1)`, messageID,
	).Scan(&exists)
	suite.Require().NoError(err)
	suite.False(exists, "message_id should NOT be stored on handler failure")
}

func TestIdempotentConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(IdempotentConsumerTestSuite))
}

func getMigrationPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(workingDir, "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")
	return fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
}

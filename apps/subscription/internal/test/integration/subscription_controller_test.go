//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"subscription-service/internal/application/event"
	"subscription-service/internal/application/services/subscription"
	"subscription-service/internal/config"
	"subscription-service/internal/domain"
	appPostgres "subscription-service/internal/infrastructure/db/postgres"
	"subscription-service/internal/infrastructure/db/postgres/outbox"
	pgevent "subscription-service/internal/infrastructure/db/postgres/outbox/event"
	pgsubscription "subscription-service/internal/infrastructure/db/postgres/subscription"
	rbevent "subscription-service/internal/infrastructure/rabbitmq/event"
	"subscription-service/internal/interface/rest"
	"subscription-service/internal/test/containers"
	"subscription-service/internal/test/stubs"
	"subscription-service/pkg/middleware"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type SubscriptionControllerTestSuite struct {
	suite.Suite
	Router   *gin.Engine
	Postgres *containers.PostgresContainer
	DB       *pgxpool.Pool
}

func (suite *SubscriptionControllerTestSuite) SetupSuite() {
	ctx := context.Background()
	serverHost := "http://localhost:8080"

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
	suite.DB = db.Pool
	appPostgres.RunMigrationsWithPath(cfg, getMigrationPath())

	cityValidator := stubs.NewCityValidatorStub()
	recorder := stubs.NewRecorderStub()
	outboxRepo := outbox.NewOutboxRepository(db.Pool)
	dispatcher := event.NewDispatcher()
	dispatcher.Register(pgevent.NewUserSubscribedHandler(serverHost, outboxRepo))
	dispatcher.Register(rbevent.NewWeatherUpdateHandler(serverHost, stubs.NewPublisherStub()))
	subscriptionRepo := pgsubscription.NewRepository(db.Pool)
	subscriptionService := subscription.NewService(
		subscriptionRepo, cityValidator, dispatcher, recorder,
	)
	subscriptionController := rest.NewSubscriptionController(subscriptionService)
	txManager := middleware.NewTxManager(db.Pool)

	router := gin.Default()
	router.Use(middleware.HttpErrorHandler())
	router.Use(middleware.HttpTransaction(txManager))

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionController.Subscribe)
		api.GET("/confirm/:token", subscriptionController.Confirm)
		api.GET("/unsubscribe/:token", subscriptionController.Unsubscribe)
	}

	suite.Router = router
}

func (suite *SubscriptionControllerTestSuite) TearDownTest() {
	_, err := suite.DB.Exec(context.Background(), "DELETE FROM subscriptions")
	suite.Require().NoError(err)
}

func (suite *SubscriptionControllerTestSuite) insertTestData(s *domain.Subscription) {
	ctx := context.Background()
	now := time.Now()

	_, err := suite.DB.Exec(ctx, `
		INSERT INTO subscriptions (
			email, city, frequency, token, confirmed, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		s.Email,
		s.City,
		s.Frequency,
		s.Token,
		s.Confirmed,
		now,
		now,
	)
	suite.Require().NoError(err)
}

func (suite *SubscriptionControllerTestSuite) TestSubscribe() {
	formData := "email=test@example.com&city=London&frequency=daily"
	body := strings.NewReader(formData)

	req, reqErr := http.NewRequest(http.MethodPost, "/api/subscribe", body)
	suite.Require().NoError(reqErr)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusOK, resp.Code)

	var count int64
	err := suite.DB.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM subscriptions WHERE email = $1`,
		"test@example.com",
	).Scan(&count)
	suite.Require().NoError(err)
	suite.Equal(int64(1), count)
}

func (suite *SubscriptionControllerTestSuite) TestSubscribe_InvalidInput() {
	formData := "email=test@example.com"
	body := strings.NewReader(formData)

	req, reqErr := http.NewRequest(http.MethodPost, "/api/subscribe", body)
	suite.Require().NoError(reqErr)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusBadRequest, resp.Code)
	suite.Contains(resp.Body.String(), "Invalid input")
}

func (suite *SubscriptionControllerTestSuite) TestSubscribe_EmailAlreadySubscribed() {
	formData := "email=test@example.com&city=London&frequency=daily"
	body := strings.NewReader(formData)

	req1, reqErr := http.NewRequest(http.MethodPost, "/api/subscribe", body)
	suite.Require().NoError(reqErr)
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp1 := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp1, req1)
	suite.Equal(http.StatusOK, resp1.Code)

	bodyDuplicate := strings.NewReader(formData)
	req2, req2Err := http.NewRequest(http.MethodPost, "/api/subscribe", bodyDuplicate)
	suite.Require().NoError(req2Err)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2 := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp2, req2)

	suite.Equal(http.StatusConflict, resp2.Code)
	suite.Contains(resp2.Body.String(), "Email already subscribed")
}

func (suite *SubscriptionControllerTestSuite) TestConfirmSubscription() {
	token := "test-token"
	suite.insertTestData(&domain.Subscription{
		Email:     "test@example.com",
		City:      "London",
		Frequency: "daily",
		Token:     token,
		Confirmed: false,
	})

	req, reqErr := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/confirm/%s", token), nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Contains(response, "message")

	var confirmed bool
	err = suite.DB.QueryRow(
		context.Background(),
		`SELECT confirmed FROM subscriptions WHERE token = $1`,
		token,
	).Scan(&confirmed)
	suite.Require().NoError(err)
	suite.True(confirmed)
}

func (suite *SubscriptionControllerTestSuite) TestConfirmSubscription_InvalidToken() {
	req, reqErr := http.NewRequest(http.MethodGet, "/api/confirm/ ", nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusBadRequest, resp.Code)
	suite.Contains(resp.Body.String(), "Invalid token")
}

func (suite *SubscriptionControllerTestSuite) TestConfirmSubscription_TokenNotFound() {
	nonExistentToken := "non-existent-token"
	req, reqErr := http.NewRequest(
		http.MethodGet, fmt.Sprintf("/api/confirm/%s", nonExistentToken), nil,
	)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusNotFound, resp.Code)
	suite.Contains(resp.Body.String(), "Token not found")
}

func (suite *SubscriptionControllerTestSuite) TestUnsubscribe() {
	token := "test-token"
	suite.insertTestData(&domain.Subscription{
		Email:     "test@example.com",
		City:      "London",
		Frequency: "daily",
		Token:     token,
		Confirmed: true,
	})

	req, reqErr := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/unsubscribe/%s", token), nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Contains(response, "message")

	var count int64
	err = suite.DB.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM subscriptions WHERE token = $1`,
		token,
	).Scan(&count)
	suite.Require().NoError(err)
	suite.Equal(int64(0), count)
}

func (suite *SubscriptionControllerTestSuite) TestUnsubscribe_InvalidToken() {
	req, reqErr := http.NewRequest(http.MethodGet, "/api/unsubscribe/ ", nil)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusBadRequest, resp.Code)
	suite.Contains(resp.Body.String(), "Invalid token")
}

func (suite *SubscriptionControllerTestSuite) TestUnsubscribe_TokenNotFound() {
	nonExistentToken := "non-existent-token"
	req, reqErr := http.NewRequest(http.MethodGet, fmt.Sprintf(
		"/api/unsubscribe/%s", nonExistentToken), nil,
	)
	suite.Require().NoError(reqErr)
	resp := httptest.NewRecorder()
	suite.Router.ServeHTTP(resp, req)

	suite.Equal(http.StatusNotFound, resp.Code)
	suite.Contains(resp.Body.String(), "Token not found")
}

func TestSubscriptionControllerTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionControllerTestSuite))
}

func getMigrationPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("getting working directory")
	}
	projectRoot := filepath.Join(workingDir, "../../..")
	migrationsPath := filepath.Join(projectRoot, "migrations")
	return fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
}

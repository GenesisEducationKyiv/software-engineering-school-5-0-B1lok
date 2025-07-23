package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"subscription-service/internal/application/event"
	"subscription-service/internal/infrastructure/db/postgres/outbox"
	pgevent "subscription-service/internal/infrastructure/db/postgres/outbox/event"
	pgsubscription "subscription-service/internal/infrastructure/db/postgres/subscription"
	rbevent "subscription-service/internal/infrastructure/rabbitmq/event"

	"subscription-service/internal/application/scheduled"
	"subscription-service/internal/application/services/subscription"
	"subscription-service/internal/config"
	"subscription-service/internal/infrastructure/db/postgres"
	"subscription-service/internal/infrastructure/grpc/validator"
	"subscription-service/internal/infrastructure/rabbitmq"
	"subscription-service/internal/interface/rest"
	"subscription-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("application failed to start")
	}
}

//nolint:gocyclo
func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	postgres.RunMigrations(cfg.DB)

	db, err := postgres.ConnectDB(cfg.DB)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize infrastructure components
	txManager := middleware.NewTxManager(db)
	validationClient, conn, err := validator.NewClient(cfg.ValidatorAddress)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create validation client: %w", err)
	}
	defer func() {
		if conn != nil {
			if closeErr := conn.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("failed to close connection")
			}
		}
	}()
	cityValidator := validator.NewCityValidator(validationClient)

	rabbitConn, err := rabbitmq.NewConnection(cfg.RabbitMqURL)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer func() {
		if err := rabbitConn.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close connection")
		}
	}()

	rabbitmqChannel, err := rabbitmq.NewChannel(rabbitConn)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}
	defer func() {
		if err := rabbitmqChannel.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close RabbitMQ channel")
		}
	}()

	if err := rabbitmq.DeclareQueues(rabbitmqChannel, rabbitmq.GetAppQueueConfigs()); err != nil {
		cancel()
		return fmt.Errorf("failed to declare RabbitMQ queues: %w", err)
	}

	// Initialize repositories
	subscriptionRepo := pgsubscription.NewRepository(db)
	outboxRepo := outbox.NewOutboxRepository(db)

	// Initialize services
	publisher, err := rabbitmq.NewPublisher(rabbitmqChannel)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}
	dispatcher := event.NewDispatcher()
	dispatcher.Register(pgevent.NewUserSubscribedHandler(cfg.Server.Host, outboxRepo))
	dispatcher.Register(rbevent.NewWeatherUpdateHandler(cfg.Server.Host, publisher))
	subscriptionService := subscription.NewService(
		subscriptionRepo, cityValidator, dispatcher)

	// Initialize controllers
	subscriptionController := rest.NewSubscriptionController(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		subscriptionRepo, dispatcher))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		subscriptionRepo, dispatcher))
	go jobManager.StartScheduler()

	// Initialize router
	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.Use(middleware.HttpErrorHandler())
	router.Use(middleware.HttpTransaction(txManager))

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionController.Subscribe)
		api.GET("/confirm/:token", subscriptionController.Confirm)
		api.GET("/unsubscribe/:token", subscriptionController.Unsubscribe)
	}

	go func() {
		<-ctx.Done()
		cancel()
	}()

	serverAddr := fmt.Sprintf(":%s", cfg.Server.HttpPort)
	log.Info().Str("address", serverAddr).Msg("Starting HTTP server")
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

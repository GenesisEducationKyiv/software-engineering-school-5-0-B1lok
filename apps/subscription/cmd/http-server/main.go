package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
		log.Fatalf("Application failed to start: %v", err)
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
				log.Printf("Failed to close connection: %v", closeErr)
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
			log.Printf("Failed to close RabbitMQ connection: %v", err)
		}
	}()

	rabbitmqChannel, err := rabbitmq.NewChannel(rabbitConn)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}
	defer func() {
		if err := rabbitmqChannel.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ channel: %v", err)
		}
	}()

	if err := rabbitmq.DeclareQueues(rabbitmqChannel); err != nil {
		cancel()
		return fmt.Errorf("failed to declare RabbitMQ queues: %w", err)
	}

	// Initialize repositories
	subscriptionRepo := postgres.NewSubscriptionRepository(db)

	// Initialize services
	publisher := rabbitmq.NewPublisher(rabbitmqChannel, cfg.Server.Host)
	subscriptionService := subscription.NewService(
		subscriptionRepo, cityValidator, publisher)

	// Initialize controllers
	subscriptionController := rest.NewSubscriptionController(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		subscriptionRepo, publisher))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		subscriptionRepo, publisher))
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
	log.Printf("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

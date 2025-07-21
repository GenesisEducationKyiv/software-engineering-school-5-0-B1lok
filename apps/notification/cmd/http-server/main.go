package main

import (
	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"notification/internal/application/event"
	"notification/internal/infrastructure/email"
	infevent "notification/internal/infrastructure/event"
	grpcweather "notification/internal/infrastructure/grpc/weather"
	"notification/internal/interface/rabbitmq"
	"notification/internal/postgres"
	"notification/pkg"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/B1lok/proto-contracts"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"notification/internal/config"
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

	db, err := postgres.ConnectDB(ctx, cfg.DB)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

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

	if err := rabbitmq.DeclareQueues(rabbitmqChannel, rabbitmq.GetAppQueueConfigs()); err != nil {
		cancel()
		return fmt.Errorf("failed to declare RabbitMQ queues: %w", err)
	}

	// Initialize the Weather Service client
	weatherConn, err := grpc.NewClient(
		cfg.WeatherServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer func() {
		if weatherConn != nil {
			if closeErr := weatherConn.Close(); closeErr != nil {
				log.Printf("Failed to close connection: %v", closeErr)
			}
		}
	}()
	grpcWeatherClient := grpcweather.NewWeatherServiceClient(weatherConn)
	appWeatherClient := grpcweather.NewClient(grpcWeatherClient)
	repository := postgres.NewRepository(db.Pool)
	txManager := pkg.NewTxManager(db.Pool)

	consumer := rabbitmq.NewConsumer(rabbitmqChannel)
	idempotentConsumer := rabbitmq.NewIdempotentConsumer(rabbitmqChannel, txManager, repository)
	sender := email.NewEmailSender(cfg.Email)

	dispatcher := event.NewDispatcher(consumer)
	idempotentDispatcher := event.NewDispatcher(idempotentConsumer)
	userSubscribedHandler := infevent.NewUserSubscribedHandler(sender)
	weatherUpdateHandler := infevent.NewWeatherUpdateHandler(appWeatherClient, sender)
	if err := idempotentDispatcher.Register(ctx, userSubscribedHandler); err != nil {
		cancel()
		return fmt.Errorf("failed to register user subscribed handler: %w", err)
	}
	if err := dispatcher.Register(ctx, weatherUpdateHandler); err != nil {
		cancel()
		return fmt.Errorf("failed to register weather updated handler: %w", err)
	}

	router := gin.Default()

	go func() {
		<-ctx.Done()
		cancel()
	}()

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/B1lok/proto-contracts"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"

	"subscription-service/internal/application/scheduled"
	"subscription-service/internal/application/services/subscription"
	"subscription-service/internal/config"
	"subscription-service/internal/infrastructure/db/postgres"
	"subscription-service/internal/infrastructure/grpc/validator"
	"subscription-service/internal/infrastructure/rabbitmq"
	grpcsubscription "subscription-service/internal/interface/grpc/subscription"
	"subscription-service/pkg/middleware"
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
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close connection: %v", closeErr)
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

	// Initialize handlers
	subscriptionHandler := grpcsubscription.NewHandler(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		subscriptionRepo, publisher))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		subscriptionRepo, publisher))
	go jobManager.StartScheduler()

	// Initialize server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.GRPCErrorInterceptor(),
			middleware.GRPCTransaction(txManager),
		),
	)

	grpcsubscription.RegisterSubscriptionServiceServer(s, subscriptionHandler)

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

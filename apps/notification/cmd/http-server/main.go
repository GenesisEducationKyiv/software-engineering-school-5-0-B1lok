package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/B1lok/proto-contracts"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"notification/internal/config"
	"notification/internal/grpc/weather"
	"notification/internal/rabbitmq"
	"notification/internal/rabbitmq/consumer"
	"notification/internal/rabbitmq/publisher"
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
	grpcWeatherClient := weather.NewWeatherServiceClient(weatherConn)
	appWeatherClient := weather.NewClient(grpcWeatherClient)

	emailPublisher := publisher.NewPublisher(rabbitmqChannel, appWeatherClient)
	worker := consumer.NewWorker(rabbitmqChannel, emailPublisher)
	if err := worker.StartConfirmationConsumer(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start confirmation worker: %w", err)
	}
	if err := worker.StartHourlyUpdateConsumer(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start hourly update worker: %w", err)
	}
	if err := worker.StartDailyUpdateConsumer(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to start daily update worker: %w", err)
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

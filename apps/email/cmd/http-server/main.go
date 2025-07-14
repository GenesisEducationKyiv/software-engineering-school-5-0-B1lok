package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"email/internal/config"
	"email/internal/email"
	"email/internal/rabbitmq"
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

	emailSender := email.NewEmailSender(cfg.Email)
	consumer := rabbitmq.NewConsumer(rabbitmqChannel, emailSender)
	if err := consumer.StartConfirmationConsumer(); err != nil {
		cancel()
		return fmt.Errorf("failed to start confirmation worker: %w", err)
	}
	if err := consumer.StartHourlyUpdateConsumer(); err != nil {
		cancel()
		return fmt.Errorf("failed to start hourly update worker: %w", err)
	}
	if err := consumer.StartDailyUpdateConsumer(); err != nil {
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

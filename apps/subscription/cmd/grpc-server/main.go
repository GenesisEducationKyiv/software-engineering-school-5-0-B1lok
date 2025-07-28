package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"subscription-service/internal/infrastructure/prometheus"

	"github.com/rs/zerolog/log"

	"subscription-service/internal/application/event"
	"subscription-service/internal/infrastructure/db/postgres/outbox"
	pgevent "subscription-service/internal/infrastructure/db/postgres/outbox/event"
	"subscription-service/internal/infrastructure/db/postgres/outbox/relay"
	pgsubscription "subscription-service/internal/infrastructure/db/postgres/subscription"

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
	rbevent "subscription-service/internal/infrastructure/rabbitmq/event"
	grpcsubscription "subscription-service/internal/interface/grpc/subscription"
	"subscription-service/pkg/middleware"
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
	appMetrics := prometheus.NewAppMetrics()
	validationClient, conn, err := validator.NewClient(cfg.ValidatorAddress)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create validation client: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close validation client connection")
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
			log.Error().Err(err).Msg("failed to close RabbitMQ connection")
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
		subscriptionRepo, cityValidator, dispatcher, appMetrics)

	// Initialize handlers
	subscriptionHandler := grpcsubscription.NewHandler(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		subscriptionRepo, dispatcher))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		subscriptionRepo, dispatcher))
	jobManager.RegisterJob(relay.NewRelayJob(publisher, txManager, outboxRepo))
	go jobManager.StartScheduler()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.MetricsInterceptor(appMetrics),
			middleware.GRPCErrorInterceptor(),
			middleware.GRPCTransaction(txManager),
		),
	)

	grpcsubscription.RegisterSubscriptionServiceServer(s, subscriptionHandler)

	// Initialize server
	grpcErrChan := make(chan error, 1)
	httpErrChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.MetricsPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	go func() {
		log.Info().Msgf("gRPC server listening on :%s", cfg.Server.GrpcPort)
		if err := s.Serve(lis); err != nil {
			grpcErrChan <- err
		}
	}()

	go func() {
		log.Info().Msgf("Metrics HTTP server listening on :%s", cfg.MetricsPort)
		if err := metricsServer.ListenAndServe(); err != nil {
			httpErrChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutdown initiated...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Metrics HTTP shutdown error")
		}
		s.GracefulStop()
		log.Info().Msg("Gracefully stopped")
		return nil

	case err := <-httpErrChan:
		log.Printf("HTTP metrics server error: %v", err)
		return err

	case err := <-grpcErrChan:
		log.Printf("gRPC server error: %v", err)
		return err
	}
}

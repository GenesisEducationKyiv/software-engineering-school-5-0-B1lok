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

	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	weatherservice "weather-service/internal/application/services/weather"
	"weather-service/internal/config"
	"weather-service/internal/infrastructure"
	"weather-service/internal/infrastructure/db/redis"
	cacheValidator "weather-service/internal/infrastructure/db/redis/validator"
	cacheClient "weather-service/internal/infrastructure/db/redis/weather"
	"weather-service/internal/infrastructure/db/redis/weather/ttl"
	"weather-service/internal/infrastructure/http/validator"
	"weather-service/internal/infrastructure/http/validator/providers/geocoding"
	client "weather-service/internal/infrastructure/http/validator/providers/weather-api-search"
	"weather-service/internal/infrastructure/http/weather"
	openmeteo "weather-service/internal/infrastructure/http/weather/providers/open-meteo"
	weatherapi "weather-service/internal/infrastructure/http/weather/providers/weather-api"
	"weather-service/internal/infrastructure/prometheus"
	grpcvalidator "weather-service/internal/interface/grpc/validator"
	grpcweather "weather-service/internal/interface/grpc/weather"
	"weather-service/pkg/logger"
	"weather-service/pkg/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("application failed to start")
	}
}

//nolint:gocyclo
func run() error {
	cfg, err := config.LoadConfig()
	logger.Configure(cfg)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	redisClient, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Initialize logger
	fileLogger, err := logger.NewFileLogger("logs/weather-api.log")
	if err != nil {
		cancel()
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Initialize infrastructure components
	validatorMetrics := prometheus.NewCacheMetrics("weather-api", "validator")
	weatherMetrics := prometheus.NewCacheMetrics("weather-api", "weather")

	geoCodingApiClient := geocoding.NewClient(cfg.GeoCodingURL, fileLogger)
	cachedGeoCodingClient := cacheValidator.NewProxyClient(
		geoCodingApiClient,
		redisClient,
		24*time.Hour,
		"geo-coding",
		validatorMetrics,
	)

	weatherApiSearchClient := client.NewClient(
		cfg.Weather.ApiURL, cfg.Weather.ApiKey, fileLogger)
	cachedWeatherApiSearchClient := cacheValidator.NewProxyClient(
		weatherApiSearchClient,
		redisClient,
		24*time.Hour,
		"weather-search",
		validatorMetrics,
	)

	geoCodingApiHandler := validator.NewHandler(cachedGeoCodingClient)
	geoCodingApiHandler.SetNext(cachedWeatherApiSearchClient)
	cityValidator := validator.NewCityValidator(geoCodingApiHandler)

	// Initialize repositories
	weatherApiClient := weatherapi.NewClient(
		cfg.Weather.ApiURL,
		cfg.Weather.ApiKey,
		fileLogger,
		infrastructure.SystemClock{},
	)
	cachedWeatherApiClient := cacheClient.NewProxyClient(
		weatherApiClient,
		redisClient,
		ttl.NewTTLProvider(15*time.Minute, infrastructure.SystemClock{}),
		"weather-api",
		weatherMetrics,
	)

	openMeteoApiClient := openmeteo.NewClient(
		cfg.OpenMeteoURL,
		cfg.GeoCodingURL,
		fileLogger,
		infrastructure.SystemClock{},
	)
	cachedOpenMeteoApi := cacheClient.NewProxyClient(
		openMeteoApiClient,
		redisClient,
		ttl.NewTTLProvider(1*time.Hour, infrastructure.SystemClock{}),
		"open-meteo",
		weatherMetrics,
	)

	openMeteoApiHandler := weather.NewHandler(cachedOpenMeteoApi)
	openMeteoApiHandler.SetNext(cachedWeatherApiClient)
	weatherRepository := weather.NewRepository(openMeteoApiHandler)

	// Initialize services
	weatherService := weatherservice.NewService(weatherRepository)

	// Initialize controllers
	weatherHandler := grpcweather.NewHandler(weatherService)
	validatorHandler := grpcvalidator.NewHandler(cityValidator)

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcErrorInterceptor()),
	)

	grpcweather.RegisterWeatherServiceServer(s, weatherHandler)
	grpcvalidator.RegisterCityValidationServiceServer(s, validatorHandler)

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

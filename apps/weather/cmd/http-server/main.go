package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	restvalidator "weather-service/internal/interface/rest/validator"
	restweather "weather-service/internal/interface/rest/weather"
	"weather-service/pkg/logger"
	"weather-service/pkg/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed to start: %v", err)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
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
	weatherController := restweather.NewController(weatherService)
	validatorController := restvalidator.NewController(cityValidator)

	// Initialize router
	router := gin.Default()

	router.Use(middleware.HttpErrorHandler())

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api")
	{
		api.GET("/weather/current", weatherController.GetWeather)
		api.GET("/weather/daily", weatherController.GetDailyForecast)
		api.GET("/weather/hourly", weatherController.GetHourlyForecast)
		api.GET("/cities/validate", validatorController.ValidateCity)
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

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

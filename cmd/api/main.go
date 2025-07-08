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

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"weather-api/internal/infrastructure/prometheus"

	geocodingapi "weather-api/internal/infrastructure/http/validator/providers/geo-coding-api"
	weatherapisearch "weather-api/internal/infrastructure/http/validator/providers/weather-api-search"
	weatherapi "weather-api/internal/infrastructure/http/weather/providers/weather-api"
	"weather-api/pkg/logger"

	"weather-api/internal/application/services/subscription"
	appWeather "weather-api/internal/application/services/weather"
	"weather-api/internal/infrastructure/http/weather"
	openmeteo "weather-api/internal/infrastructure/http/weather/providers/open-meteo"

	appEmail "weather-api/internal/application/email"
	"weather-api/internal/application/scheduled"
	"weather-api/internal/config"
	postgresconnector "weather-api/internal/infrastructure/db/postgres"
	"weather-api/internal/infrastructure/db/redis"
	cacheValidator "weather-api/internal/infrastructure/db/redis/validator"
	cacheClient "weather-api/internal/infrastructure/db/redis/weather"
	cacheOpenmeteo "weather-api/internal/infrastructure/db/redis/weather/providers/open-meteo"
	cacheWeather "weather-api/internal/infrastructure/db/redis/weather/providers/weather-api"
	"weather-api/internal/infrastructure/email"
	"weather-api/internal/infrastructure/http/validator"
	"weather-api/internal/interface/rest"
	"weather-api/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	postgresconnector.RunMigrations(cfg.DB)

	db, err := postgresconnector.ConnectDB(cfg.DB)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to database: %w", err)
	}

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
	emailSender := email.NewEmailSender(email.CreateConfig(cfg.Email))
	txManager := middleware.NewTxManager(db)

	validatorMetrics := prometheus.NewCacheMetrics("weather-api", "validator")
	weatherMetrics := prometheus.NewCacheMetrics("weather-api", "weather")

	geoCodingApiClient := geocodingapi.NewClient(cfg.GeoCodingUrl, fileLogger)
	cachedGeoCodingClient := cacheValidator.NewProxyClient(
		geoCodingApiClient,
		redisClient,
		24*time.Hour,
		"geo-coding",
		validatorMetrics,
	)

	weatherApiSearchClient := weatherapisearch.NewClient(
		cfg.Weather.ApiUrl, cfg.Weather.ApiKey, fileLogger)
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
	weatherApiClient := weatherapi.NewClient(cfg.Weather.ApiUrl, cfg.Weather.ApiKey, fileLogger)
	cachedWeatherApiClient := cacheClient.NewProxyClient(
		weatherApiClient,
		redisClient,
		cacheWeather.NewTTLProvider(),
		"weather-api",
		weatherMetrics,
	)

	openMeteoApiClient := openmeteo.NewClient(cfg.OpenMeteoUrl, cfg.GeoCodingUrl, fileLogger)
	cachedOpenMeteoApi := cacheClient.NewProxyClient(
		openMeteoApiClient,
		redisClient,
		cacheOpenmeteo.NewTTLProvider(),
		"open-meteo",
		weatherMetrics,
	)

	openMeteoApiHandler := weather.NewHandler(cachedOpenMeteoApi)
	openMeteoApiHandler.SetNext(cachedWeatherApiClient)
	weatherRepository := weather.NewRepository(openMeteoApiHandler)
	subscriptionRepo := postgresconnector.NewSubscriptionRepository(db)

	// Initialize services
	emailNotifier := appEmail.NewNotifier(cfg.Server.Host, emailSender)
	weatherService := appWeather.NewService(weatherRepository)
	subscriptionService := subscription.NewService(
		subscriptionRepo, cityValidator, emailNotifier, cfg.Server.Host)

	// Initialize controllers
	weatherController := rest.NewWeatherController(weatherService)
	subscriptionController := rest.NewSubscriptionController(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		weatherRepository, subscriptionRepo, emailNotifier))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		weatherRepository, subscriptionRepo, emailNotifier))
	go jobManager.StartScheduler()

	// Initialize router
	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.TransactionMiddleware(txManager))

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api")
	{
		api.GET("/weather", weatherController.GetWeather)
		api.POST("/subscribe", subscriptionController.Subscribe)
		api.GET("/confirm/:token", subscriptionController.Confirm)
		api.GET("/unsubscribe/:token", subscriptionController.Unsubscribe)
	}

	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	go func() {
		<-ctx.Done()
		cancel()
	}()

	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

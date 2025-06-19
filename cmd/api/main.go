package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	appEmail "weather-api/internal/application/email"
	"weather-api/internal/application/scheduled"
	"weather-api/internal/application/services"
	"weather-api/internal/config"
	postgresconnector "weather-api/internal/infrastructure/db/postgres"
	"weather-api/internal/infrastructure/email"
	cityValidator "weather-api/internal/infrastructure/http/validator"
	weatherapi "weather-api/internal/infrastructure/http/weather-api"
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
	postgresconnector.RunMigrations(cfg)

	db, err := postgresconnector.ConnectDB(cfg)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize infrastructure components
	emailSender := email.NewEmailSender(email.CreateConfig(cfg))
	txManager := middleware.NewTxManager(db)
	cityValidatorImpl := cityValidator.NewCityValidator(cfg.WeatherApiUrl, cfg.WeatherApiKey)

	// Initialize repositories
	weatherRepo := weatherapi.NewWeatherRepository(cfg.WeatherApiUrl, cfg.WeatherApiKey)
	subscriptionRepo := postgresconnector.NewSubscriptionRepository(db)

	// Initialize services
	weatherService := services.NewWeatherService(weatherRepo)
	subscriptionService := services.NewSubscriptionService(
		subscriptionRepo, cityValidatorImpl, emailSender, cfg.ServerHost)
	emailNotifier := appEmail.NewNotifier(cfg.ServerHost, emailSender)

	// Initialize controllers
	weatherController := rest.NewWeatherController(weatherService)
	subscriptionController := rest.NewSubscriptionController(subscriptionService)

	// Initialize workers
	jobManager := scheduled.NewJobManager(ctx)
	jobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		weatherRepo, subscriptionRepo, emailNotifier))
	jobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		weatherRepo, subscriptionRepo, emailNotifier))
	go jobManager.StartScheduler()

	// Initialize router
	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.TransactionMiddleware(txManager))

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	api := router.Group("/api")
	{
		api.GET("/weather", weatherController.GetWeather)
		api.POST("/subscribe", subscriptionController.Subscribe)
		api.GET("/confirm/:token", subscriptionController.Confirm)
		api.GET("/unsubscribe/:token", subscriptionController.Unsubscribe)
	}

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

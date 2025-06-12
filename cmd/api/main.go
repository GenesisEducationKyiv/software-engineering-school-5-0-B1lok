package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"weather-api/internal/application/scheduled"
	"weather-api/internal/application/services"
	"weather-api/internal/config"
	postgresconnector "weather-api/internal/infrastructure/db/postgres"
	"weather-api/internal/infrastructure/email"
	cityValidator "weather-api/internal/infrastructure/http/validator"
	weatherapi "weather-api/internal/infrastructure/http/weather-api"
	"weather-api/internal/interface/api/rest"
	"weather-api/pkg/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

type Components struct {
	WeatherRepo            *weatherapi.WeatherRepository
	WeatherService         *services.WeatherService
	WeatherController      *rest.WeatherController
	CityValidator          *cityValidator.CityValidator
	SubscriptionRepo       *postgresconnector.SubscriptionRepository
	SubscriptionService    *services.SubscriptionService
	SubscriptionController *rest.SubscriptionController
	Sender                 *email.Sender
	TxManager              middleware.TxManager
	JobManager             *scheduled.JobManager
}

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

	components := setupComponents(ctx, cfg, db)
	setupScheduledJobs(components, cfg)

	router := createRouter(components)

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

func setupComponents(ctx context.Context, cfg config.Config, db *gorm.DB) *Components {
	weatherRepo := weatherapi.NewWeatherRepository(cfg.WeatherApiKey)
	weatherService := services.NewWeatherService(weatherRepo)
	weatherController := rest.NewWeatherController(weatherService)
	cityValidatorImpl := cityValidator.NewCityValidator(cfg.WeatherApiKey)
	sender := email.NewEmailSender(email.CreateConfig(cfg))
	txManager := middleware.NewTxManager(db)

	subscriptionRepo := postgresconnector.NewSubscriptionRepository(db)
	subscriptionService := services.NewSubscriptionService(
		subscriptionRepo, cityValidatorImpl, sender, cfg.ServerHost)
	subscriptionController := rest.NewSubscriptionController(subscriptionService)

	jm := scheduled.NewJobManager(ctx)
	return &Components{
		WeatherRepo:            weatherRepo,
		WeatherService:         weatherService,
		WeatherController:      weatherController,
		CityValidator:          cityValidatorImpl,
		SubscriptionRepo:       subscriptionRepo,
		SubscriptionService:    subscriptionService,
		SubscriptionController: subscriptionController,
		Sender:                 sender,
		TxManager:              txManager,
		JobManager:             jm,
	}
}

func setupScheduledJobs(components *Components, cfg config.Config) {
	components.JobManager.RegisterJob(scheduled.NewHourlyWeatherUpdateJob(
		components.WeatherRepo, components.SubscriptionRepo, components.Sender, cfg.ServerHost))
	components.JobManager.RegisterJob(scheduled.NewDailyWeatherUpdateJob(
		components.WeatherRepo, components.SubscriptionRepo, components.Sender, cfg.ServerHost))
	go components.JobManager.StartScheduler()
}

func createRouter(components *Components) *gin.Engine {
	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.Use(middleware.ErrorHandler())
	router.Use(middleware.TransactionMiddleware(components.TxManager))

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	api := router.Group("/api")
	{
		api.GET("/weather", components.WeatherController.GetWeather)
		api.POST("/subscribe", components.SubscriptionController.Subscribe)
		api.GET("/confirm/:token", components.SubscriptionController.Confirm)
		api.GET("/unsubscribe/:token", components.SubscriptionController.Unsubscribe)
	}

	return router
}

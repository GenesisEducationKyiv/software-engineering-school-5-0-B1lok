package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gateway/internal/config"
	"gateway/internal/controllers/subscription"
	"gateway/internal/controllers/weather"
	"gateway/pkg/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Application failed to start")
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize the Subscription Service client
	subscriptionConn, err := grpc.NewClient(
		cfg.SubscriptionServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer func() {
		if subscriptionConn != nil {
			if closeErr := subscriptionConn.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Failed to close gRPC connection")
			}
		}
	}()
	subscriptionClient := subscription.NewSubscriptionServiceClient(subscriptionConn)

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
				log.Error().Err(closeErr).Msg("Failed to close gRPC connection")
			}
		}
	}()
	weatherClient := weather.NewWeatherServiceClient(weatherConn)

	// Initialize controllers
	subscriptionController := subscription.NewController(subscriptionClient)
	weatherController := weather.NewController(weatherClient)

	// Initialize router
	router := gin.Default()

	router.LoadHTMLGlob("templates/index.html")

	router.Use(middleware.HttpErrorHandler())

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionController.Subscribe)
		api.GET("/confirm/:token", subscriptionController.Confirm)
		api.GET("/unsubscribe/:token", subscriptionController.Unsubscribe)

		api.GET("/weather/current", weatherController.GetWeather)
		api.GET("/weather/daily", weatherController.GetDailyForecast)
		api.GET("/weather/hourly", weatherController.GetHourlyForecast)
	}

	go func() {
		<-ctx.Done()
		cancel()
	}()

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Info().Str("address", serverAddr).Msg("Starting HTTP server")
	if err := router.Run(serverAddr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"notification/internal/application/event"
	"notification/internal/infrastructure/grpc/weather"
)

const (
	weatherDailyTemplate  = "daily.html"
	weatherHourlyTemplate = "hourly.html"
)

type WeatherClient interface {
	DailyUpdate(ctx context.Context, city string) (*weather.WeatherDaily, error)
	HourlyUpdate(ctx context.Context, city string) (*weather.WeatherHourly, error)
}

type WeatherUpdatedHandler struct {
	client WeatherClient
	sender Sender
}

type weatherUpdatedPayload struct {
	Email          string `json:"email"`
	City           string `json:"city"`
	Frequency      string `json:"frequency"`
	UnsubscribeURL string `json:"unsubscribe_url"`
}

type weatherUpdatedTemplate struct {
	UnsubscribeURL string
	Frequency      string
	Weather        any
}

func NewWeatherUpdateHandler(client WeatherClient, sender Sender) *WeatherUpdatedHandler {
	return &WeatherUpdatedHandler{
		client: client,
		sender: sender,
	}
}

func (h *WeatherUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var data weatherUpdatedPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	template, weatherData, err := h.getWeatherData(ctx, data.City, data.Frequency)

	if err != nil {
		return fmt.Errorf("failed to get weather data: %w", err)
	}

	err = h.sender.Send(
		template,
		data.Email,
		fmt.Sprintf("Your weather %s forecast", data.Frequency),
		&weatherUpdatedTemplate{
			UnsubscribeURL: data.UnsubscribeURL,
			Frequency:      data.Frequency,
			Weather:        weatherData,
		})

	if err != nil {
		return fmt.Errorf("failed to send weather update email: %w", err)
	}

	return nil
}

func (h *WeatherUpdatedHandler) GetName() event.Name {
	return event.WeatherUpdatedEventName
}

func (h *WeatherUpdatedHandler) getWeatherData(
	ctx context.Context,
	city,
	frequency string,
) (template string, data any, err error) {
	switch frequency {
	case "daily":
		weatherData, err := h.client.DailyUpdate(ctx, city)
		if err != nil {
			return "", nil, fmt.Errorf(
				"failed to fetch daily weather: %w", err,
			)
		}
		return weatherDailyTemplate, weatherData, nil

	case "hourly":
		weatherData, err := h.client.HourlyUpdate(ctx, city)
		if err != nil {
			return "", nil, fmt.Errorf(
				"failed to fetch hourly weather: %w", err,
			)
		}
		return weatherHourlyTemplate, weatherData, nil

	default:
		return "", nil, fmt.Errorf("unsupported frequency: %s", frequency)
	}
}

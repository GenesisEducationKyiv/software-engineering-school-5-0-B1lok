package weather

import (
	"context"

	"weather-service/internal/domain"
)

type Handler struct {
	client Client
	next   Client
}

func NewHandler(client Client) *Handler {
	return &Handler{client: client}
}

func (h *Handler) SetNext(next Client) {
	h.next = next
}

func (h *Handler) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	resp, err := h.client.GetWeather(ctx, city)
	if err != nil && h.next != nil {
		return h.next.GetWeather(ctx, city)
	}
	return resp, err
}

func (h *Handler) GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error) {
	resp, err := h.client.GetDailyForecast(ctx, city)
	if err != nil && h.next != nil {
		return h.next.GetDailyForecast(ctx, city)
	}
	return resp, err
}

func (h *Handler) GetHourlyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherHourly, error) {
	resp, err := h.client.GetHourlyForecast(ctx, city)
	if err != nil && h.next != nil {
		return h.next.GetHourlyForecast(ctx, city)
	}
	return resp, err
}

package weather

import (
	"weather-api/internal/domain"
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

func (h *Handler) GetWeather(city string) (*domain.Weather, error) {
	resp, err := h.client.GetWeather(city)
	if err != nil && h.next != nil {
		return h.next.GetWeather(city)
	}
	return resp, err
}

func (h *Handler) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	resp, err := h.client.GetDailyForecast(city)
	if err != nil && h.next != nil {
		return h.next.GetDailyForecast(city)
	}
	return resp, err
}

func (h *Handler) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	resp, err := h.client.GetHourlyForecast(city)
	if err != nil && h.next != nil {
		return h.next.GetHourlyForecast(city)
	}
	return resp, err
}

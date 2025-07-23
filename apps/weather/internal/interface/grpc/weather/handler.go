package weather

import (
	"context"

	"weather-service/internal/application/query"
	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

type Service interface {
	GetWeather(ctx context.Context, city string) (*query.WeatherResult, error)
	GetDailyForecast(ctx context.Context, city string) (*query.WeatherDailyResult, error)
	GetHourlyForecast(ctx context.Context, city string) (*query.WeatherHourlyResult, error)
}

type Handler struct {
	UnimplementedWeatherServiceServer
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetCurrentWeather(
	ctx context.Context,
	request *CityRequest,
) (*Weather, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	city := request.GetCity()

	weather, err := h.service.GetWeather(ctx, city)
	if err != nil {
		return nil, err
	}

	return toProtoWeather(weather), nil
}

func (h *Handler) GetHourlyWeather(
	ctx context.Context,
	request *CityRequest,
) (*WeatherHourly, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	city := request.GetCity()

	weather, err := h.service.GetHourlyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	return toProtoWeatherHourly(weather), nil
}

func (h *Handler) GetDailyWeather(
	ctx context.Context,
	request *CityRequest,
) (*WeatherDaily, error) {
	err := request.ValidateAll()
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, err.Error())
	}
	city := request.GetCity()

	weather, err := h.service.GetDailyForecast(ctx, city)
	if err != nil {
		return nil, err
	}

	return toProtoWeatherDaily(weather), nil
}

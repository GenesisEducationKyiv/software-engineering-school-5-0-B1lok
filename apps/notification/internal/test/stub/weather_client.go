package stub

import (
	"context"

	"notification/internal/infrastructure/grpc/weather"
)

type WeatherClient struct {
	DailyUpdateFunc  func(ctx context.Context, city string) (*weather.WeatherDaily, error)
	HourlyUpdateFunc func(ctx context.Context, city string) (*weather.WeatherHourly, error)
}

func (s *WeatherClient) DailyUpdate(
	ctx context.Context,
	city string,
) (*weather.WeatherDaily, error) {
	if s.DailyUpdateFunc != nil {
		return s.DailyUpdateFunc(ctx, city)
	}
	return &weather.WeatherDaily{}, nil
}

func (s *WeatherClient) HourlyUpdate(
	ctx context.Context,
	city string,
) (*weather.WeatherHourly, error) {
	if s.HourlyUpdateFunc != nil {
		return s.HourlyUpdateFunc(ctx, city)
	}
	return &weather.WeatherHourly{}, nil
}

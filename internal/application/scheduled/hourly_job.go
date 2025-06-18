package scheduled

import (
	"context"

	"weather-api/internal/domain"
)

type WeatherHourlyNotifier interface {
	NotifyHourlyWeather(
		ctx context.Context,
		subscription *domain.Subscription,
		weatherHourly *domain.WeatherHourly,
	) error
}

type WeatherHourlyReader interface {
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
}

type HourlyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*domain.WeatherHourly]
}

func NewHourlyWeatherUpdateJob(
	weatherRepo WeatherHourlyReader,
	subscriptionRepo GroupedSubscriptionReader,
	notifier WeatherHourlyNotifier,
) *HourlyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*domain.WeatherHourly, error) {
		return weatherRepo.GetHourlyForecast(ctx, city)
	}

	notifyFunc := func(
		ctx context.Context,
		subscription *domain.Subscription,
		weatherHourly *domain.WeatherHourly,
	) error {
		return notifier.NotifyHourlyWeather(ctx, subscription, weatherHourly)
	}

	exec := NewWeatherJobExecutor(
		subscriptionRepo,
		"hourly",
		getWeatherFunc,
		notifyFunc,
	)

	return &HourlyWeatherUpdateJob{
		executor: exec,
	}
}

func (h *HourlyWeatherUpdateJob) Name() string {
	return "HourlyWeatherUpdateJob"
}

func (h *HourlyWeatherUpdateJob) Schedule() string {
	return "*/15 * * * * *"
}

func (h *HourlyWeatherUpdateJob) Run(ctx context.Context) error {
	return h.executor.Execute(ctx)
}

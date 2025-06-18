package scheduled

import (
	"context"

	"weather-api/internal/domain"
)

type WeatherDailyNotifier interface {
	NotifyDailyWeather(
		ctx context.Context,
		subscription *domain.Subscription,
		weatherDaily *domain.WeatherDaily,
	) error
}

type WeatherDailyReader interface {
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
}

type DailyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*domain.WeatherDaily]
}

func NewDailyWeatherUpdateJob(
	weatherRepo WeatherDailyReader,
	subscriptionRepo GroupedSubscriptionReader,
	notifier WeatherDailyNotifier,
) *DailyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*domain.WeatherDaily, error) {
		return weatherRepo.GetDailyForecast(ctx, city)
	}

	notifyFunc := func(
		ctx context.Context,
		subscription *domain.Subscription,
		weatherDaily *domain.WeatherDaily,
	) error {
		return notifier.NotifyDailyWeather(ctx, subscription, weatherDaily)
	}

	exec := NewWeatherJobExecutor(
		subscriptionRepo,
		"daily",
		getWeatherFunc,
		notifyFunc,
	)

	return &DailyWeatherUpdateJob{
		executor: exec,
	}
}

func (d *DailyWeatherUpdateJob) Name() string {
	return "DailyWeatherUpdateJob"
}

func (d *DailyWeatherUpdateJob) Schedule() string {
	return "0 0 8 * * *"
}

func (d *DailyWeatherUpdateJob) Run(ctx context.Context) error {
	return d.executor.Execute(ctx)
}

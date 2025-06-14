package scheduled

import (
	"context"
	"fmt"

	"weather-api/internal/application/email"
	"weather-api/internal/domain"
)

type WeatherHourlySender interface {
	WeatherHourlyEmail(email *email.WeatherHourlyEmail) error
}

type WeatherHourlyReader interface {
	GetHourlyForecast(ctx context.Context, city string) (*domain.WeatherHourly, error)
}

type HourlyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*domain.WeatherHourly, *HourlyEmailTask]
}

type HourlyEmailTask struct {
	subscription  *domain.Subscription
	weatherHourly *domain.WeatherHourly
	host          string
}

func (t *HourlyEmailTask) GetSubscription() *domain.Subscription {
	return t.subscription
}

func NewHourlyWeatherUpdateJob(
	weatherRepo WeatherHourlyReader,
	subscriptionRepo GroupedSubscriptionReader,
	sender WeatherHourlySender,
	host string,
) *HourlyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*domain.WeatherHourly, error) {
		return weatherRepo.GetHourlyForecast(ctx, city)
	}

	createTaskFunc := func(subscription *domain.Subscription,
		weather *domain.WeatherHourly,
	) *HourlyEmailTask {
		return &HourlyEmailTask{
			subscription:  subscription,
			weatherHourly: weather,
			host:          host,
		}
	}

	sendEmailFunc := func(task *HourlyEmailTask) error {
		return sender.WeatherHourlyEmail(
			toWeatherHourlyEmail(task.subscription, task.weatherHourly, task.host),
		)
	}

	exec := NewWeatherJobExecutor(
		subscriptionRepo,
		host,
		"hourly",
		getWeatherFunc,
		createTaskFunc,
		sendEmailFunc,
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

func toWeatherHourlyEmail(
	subscription *domain.Subscription,
	weatherHourly *domain.WeatherHourly, host string,
) *email.WeatherHourlyEmail {
	return &email.WeatherHourlyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		WeatherHourly:  weatherHourly,
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", host, subscription.Token),
	}
}

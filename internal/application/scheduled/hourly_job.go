package scheduled

import (
	"context"
	"fmt"

	"weather-api/internal/application/email"
	"weather-api/internal/domain/models"
	"weather-api/internal/domain/repositories"
)

type HourlyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*models.WeatherHourly, *HourlyEmailTask]
}

type HourlyEmailTask struct {
	subscription  *models.Subscription
	weatherHourly *models.WeatherHourly
	host          string
}

func (t *HourlyEmailTask) GetSubscription() *models.Subscription {
	return t.subscription
}

func NewHourlyWeatherUpdateJob(
	weatherRepo repositories.WeatherRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	sender email.Sender,
	host string,
) *HourlyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*models.WeatherHourly, error) {
		return weatherRepo.GetHourlyForecast(ctx, city)
	}

	createTaskFunc := func(subscription *models.Subscription,
		weather *models.WeatherHourly,
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
		weatherRepo,
		subscriptionRepo,
		sender,
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
	subscription *models.Subscription,
	weatherHourly *models.WeatherHourly, host string,
) *email.WeatherHourlyEmail {
	return &email.WeatherHourlyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		WeatherHourly:  weatherHourly,
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", host, subscription.Token),
	}
}

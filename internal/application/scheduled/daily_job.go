package scheduled

import (
	"context"
	"fmt"

	"weather-api/internal/application/email"
	"weather-api/internal/domain/models"
	"weather-api/internal/domain/repositories"
)

type DailyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*models.WeatherDaily, *DailyEmailTask]
}

type DailyEmailTask struct {
	subscription *models.Subscription
	weatherDaily *models.WeatherDaily
	host         string
}

func (t *DailyEmailTask) GetSubscription() *models.Subscription {
	return t.subscription
}

func NewDailyWeatherUpdateJob(
	weatherRepo repositories.WeatherRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	sender email.Sender,
	host string,
) *DailyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*models.WeatherDaily, error) {
		return weatherRepo.GetDailyForecast(ctx, city)
	}

	createTaskFunc := func(subscription *models.Subscription,
		weather *models.WeatherDaily,
	) *DailyEmailTask {
		return &DailyEmailTask{
			subscription: subscription,
			weatherDaily: weather,
			host:         host,
		}
	}

	sendEmailFunc := func(task *DailyEmailTask) error {
		return sender.WeatherDailyEmail(
			toWeatherDailyEmail(task.subscription, task.weatherDaily, task.host),
		)
	}

	exec := NewWeatherJobExecutor(
		weatherRepo,
		subscriptionRepo,
		sender,
		host,
		"daily",
		getWeatherFunc,
		createTaskFunc,
		sendEmailFunc,
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

func toWeatherDailyEmail(
	subscription *models.Subscription,
	weatherDaily *models.WeatherDaily,
	host string,
) *email.WeatherDailyEmail {
	return &email.WeatherDailyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		WeatherDaily:   weatherDaily,
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", host, subscription.Token),
	}
}

package scheduled

import (
	"context"
	"fmt"

	"weather-api/internal/application/email"
	"weather-api/internal/domain"
)

type WeatherDailySender interface {
	WeatherDailyEmail(email *email.WeatherDailyEmail) error
}

type WeatherDailyReader interface {
	GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error)
}

type DailyWeatherUpdateJob struct {
	executor *WeatherJobExecutor[*domain.WeatherDaily, *DailyEmailTask]
}

type DailyEmailTask struct {
	subscription *domain.Subscription
	weatherDaily *domain.WeatherDaily
	host         string
}

func (t *DailyEmailTask) GetSubscription() *domain.Subscription {
	return t.subscription
}

func NewDailyWeatherUpdateJob(
	weatherRepo WeatherDailyReader,
	subscriptionRepo GroupedSubscriptionReader,
	sender WeatherDailySender,
	host string,
) *DailyWeatherUpdateJob {
	getWeatherFunc := func(ctx context.Context, city string) (*domain.WeatherDaily, error) {
		return weatherRepo.GetDailyForecast(ctx, city)
	}

	createTaskFunc := func(subscription *domain.Subscription,
		weather *domain.WeatherDaily,
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
		subscriptionRepo,
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
	subscription *domain.Subscription,
	weatherDaily *domain.WeatherDaily,
	host string,
) *email.WeatherDailyEmail {
	return &email.WeatherDailyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		WeatherDaily:   weatherDaily,
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", host, subscription.Token),
	}
}

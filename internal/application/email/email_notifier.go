package email

import (
	"context"
	"fmt"

	"weather-api/internal/domain"
)

type Sender interface {
	WeatherDailyEmail(email *WeatherDailyEmail) error
	WeatherHourlyEmail(email *WeatherHourlyEmail) error
}

type Notifier struct {
	host   string
	sender Sender
}

func NewNotifier(host string, sender Sender) *Notifier {
	return &Notifier{
		host:   host,
		sender: sender,
	}
}

func (n *Notifier) NotifyDailyWeather(
	ctx context.Context,
	subscription *domain.Subscription,
	weatherDaily *domain.WeatherDaily,
) error {
	emailData := &WeatherDailyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", n.host, subscription.Token),
		WeatherDaily:   weatherDaily,
	}

	return n.sender.WeatherDailyEmail(emailData)
}

func (n *Notifier) NotifyHourlyWeather(
	ctx context.Context,
	subscription *domain.Subscription,
	weatherHourly *domain.WeatherHourly,
) error {
	emailData := &WeatherHourlyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		UnsubscribeUrl: fmt.Sprintf("%sapi/unsubscribe/%s", n.host, subscription.Token),
		WeatherHourly:  weatherHourly,
	}

	return n.sender.WeatherHourlyEmail(emailData)
}

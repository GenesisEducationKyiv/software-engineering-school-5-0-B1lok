package email

import (
	"fmt"

	"weather-api/internal/domain"
)

type Sender interface {
	WeatherDailyEmail(email *WeatherDailyEmail) error
	WeatherHourlyEmail(email *WeatherHourlyEmail) error
	ConfirmationEmail(email *ConfirmationEmail) error
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
	subscription *domain.Subscription,
	weatherDaily *domain.WeatherDaily,
) error {
	emailData := &WeatherDailyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", n.host, subscription.Token),
		WeatherDaily:   weatherDaily,
	}

	return n.sender.WeatherDailyEmail(emailData)
}

func (n *Notifier) NotifyConfirmation(subscription *domain.Subscription) error {
	confirmationEmail := &ConfirmationEmail{
		To:        subscription.Email,
		City:      subscription.City,
		Frequency: string(subscription.Frequency),
		URL:       fmt.Sprintf("%s/api/confirm/%s", n.host, subscription.Token),
	}
	if err := n.sender.ConfirmationEmail(confirmationEmail); err != nil {
		return err
	}

	return nil
}

func (n *Notifier) NotifyHourlyWeather(
	subscription *domain.Subscription,
	weatherHourly *domain.WeatherHourly,
) error {
	emailData := &WeatherHourlyEmail{
		To:             subscription.Email,
		Frequency:      string(subscription.Frequency),
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", n.host, subscription.Token),
		WeatherHourly:  weatherHourly,
	}

	return n.sender.WeatherHourlyEmail(emailData)
}

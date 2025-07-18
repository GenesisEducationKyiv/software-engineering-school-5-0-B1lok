package rabbitmq

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"subscription-service/internal/domain"
)

type Publisher struct {
	channel *amqp091.Channel
	host    string
}

func NewPublisher(channel *amqp091.Channel, serverHost string) *Publisher {
	return &Publisher{
		host:    serverHost,
		channel: channel,
	}
}

func (p *Publisher) NotifyConfirmation(subscription *domain.Subscription) error {
	msg := ConfirmationEmailMessage{
		To:        subscription.Email,
		City:      subscription.City,
		Frequency: string(subscription.Frequency),
		URL:       fmt.Sprintf("%s/api/confirm/%s", p.host, subscription.Token),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.Publish(
		"",
		confirmationQueue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *Publisher) NotifyWeatherUpdate(subscription *domain.Subscription) error {
	msg := WeatherUpdateMessage{
		To:             subscription.Email,
		City:           subscription.City,
		Frequency:      string(subscription.Frequency),
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", p.host, subscription.Token),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	queueName := weatherDailyQueue
	if subscription.Frequency == domain.FrequencyHourly {
		queueName = weatherHourlyQueue
	}

	err = p.channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"notification/internal/rabbitmq"
	"notification/internal/rabbitmq/consumer"
)

type WeatherClient interface {
	DailyUpdate(ctx context.Context, city string) (DailyUpdateTemplate, error)
	HourlyUpdate(ctx context.Context, city string) (HourlyUpdateTemplate, error)
}

type Publisher struct {
	channel *amqp091.Channel
	client  WeatherClient
}

func NewPublisher(channel *amqp091.Channel, client WeatherClient) *Publisher {
	return &Publisher{
		channel: channel,
		client:  client,
	}
}

func (p *Publisher) ConfirmationEmail(emailMsg *consumer.ConfirmationEmailMessage) error {
	msg := toConfirmationEmailMessage(emailMsg)

	if err := p.publish(rabbitmq.EmailConfirmationQueue, msg); err != nil {
		return fmt.Errorf("failed to publish hourly update message: %w", err)
	}

	return nil
}

func (p *Publisher) DailyUpdate(
	ctx context.Context,
	updateMsg *consumer.WeatherUpdateMessage,
) error {
	update, err := p.client.DailyUpdate(ctx, updateMsg.City)
	if err != nil {
		return fmt.Errorf("failed to get daily update: %w", err)
	}

	msg := toDailyUpdateMessage(updateMsg, &update)

	if err := p.publish(rabbitmq.EmailWeatherDailyQueue, msg); err != nil {
		return fmt.Errorf("failed to publish hourly update message: %w", err)
	}

	return nil
}

func (p *Publisher) HourlyUpdate(
	ctx context.Context,
	updateMsg *consumer.WeatherUpdateMessage,
) error {
	update, err := p.client.HourlyUpdate(ctx, updateMsg.City)
	if err != nil {
		return fmt.Errorf("failed to get hourly update: %w", err)
	}

	msg := toHourlyUpdateMessage(updateMsg, &update)

	if err := p.publish(rabbitmq.EmailWeatherHourlyQueue, msg); err != nil {
		return fmt.Errorf("failed to publish hourly update message: %w", err)
	}

	return nil
}

func (p *Publisher) publish(queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
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
		return fmt.Errorf("failed to publish message to queue %s: %w", queueName, err)
	}

	return nil
}

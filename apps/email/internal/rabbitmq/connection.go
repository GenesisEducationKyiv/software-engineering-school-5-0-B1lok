package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

const (
	emailConfirmationQueue  = "email.confirmation.send"
	emailWeatherHourlyQueue = "email.weather.hourly.send"
	emailWeatherDailyQueue  = "email.weather.daily.send"
)

func NewConnection(url string) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

func NewChannel(conn *amqp091.Connection) (*amqp091.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close connection: %v\n", closeErr)
		}
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	return ch, nil
}

func DeclareQueues(ch *amqp091.Channel) error {
	if err := declareQueue(ch, emailConfirmationQueue); err != nil {
		return err
	}
	if err := declareQueue(ch, emailWeatherHourlyQueue); err != nil {
		return err
	}
	if err := declareQueue(ch, emailWeatherDailyQueue); err != nil {
		return err
	}
	return nil
}

func declareQueue(ch *amqp091.Channel, name string) error {
	_, err := ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", name, err)
	}

	return nil
}

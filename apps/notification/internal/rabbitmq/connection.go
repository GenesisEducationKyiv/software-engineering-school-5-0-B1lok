package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

const (
	ConfirmationQueue  = "confirmation"
	WeatherHourlyQueue = "weather.hourly"
	WeatherDailyQueue  = "weather.daily"

	EmailConfirmationQueue  = "email.confirmation.send"
	EmailWeatherHourlyQueue = "email.weather.hourly.send"
	EmailWeatherDailyQueue  = "email.weather.daily.send"
)

var queues = []string{
	ConfirmationQueue,
	WeatherHourlyQueue,
	WeatherDailyQueue,
	EmailConfirmationQueue,
	EmailWeatherHourlyQueue,
	EmailWeatherDailyQueue,
}

var defaultQueueOptions = QueueOptions{
	Durable:    true,
	AutoDelete: false,
	Exclusive:  false,
	NoWait:     false,
}

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
	for _, q := range queues {
		if err := declareQueue(ch, q, defaultQueueOptions); err != nil {
			return err
		}
	}
	return nil
}

func declareQueue(ch *amqp091.Channel, name string, opts QueueOptions) error {
	_, err := ch.QueueDeclare(
		name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", name, err)
	}

	return nil
}

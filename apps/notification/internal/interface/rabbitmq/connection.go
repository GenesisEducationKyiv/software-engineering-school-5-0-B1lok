package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp091.Table
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

func DeclareQueues(ch *amqp091.Channel, queues []QueueConfig) error {
	for _, q := range queues {
		if _, err := ch.QueueDeclare(
			q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Args,
		); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.Name, err)
		}
	}
	return nil
}

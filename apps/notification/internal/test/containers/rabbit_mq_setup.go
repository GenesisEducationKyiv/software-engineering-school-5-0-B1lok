package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RabbitMQContainer struct {
	Container  testcontainers.Container
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
	URL        string
}

func SetupRabbitMQContainer(ctx context.Context) (*RabbitMQContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:4-management",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(60 * time.Second),
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": "test",
			"RABBITMQ_DEFAULT_PASS": "test",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start RabbitMQ container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	amqpURL := fmt.Sprintf("amqp://test:test@%s:%s/", host, port.Port())

	conn, err := amqp091.Dial(amqpURL)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ : %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		err := conn.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return &RabbitMQContainer{
		Container:  container,
		Connection: conn,
		Channel:    channel,
		URL:        amqpURL,
	}, nil
}

func (r *RabbitMQContainer) Cleanup(ctx context.Context) error {
	if r.Channel != nil {
		err := r.Channel.Close()
		if err != nil {
			return err
		}
	}
	if r.Connection != nil {
		err := r.Connection.Close()
		if err != nil {
			return err
		}
	}
	if r.Container != nil {
		return r.Container.Terminate(ctx)
	}
	return nil
}

func (r *RabbitMQContainer) DeclareQueue(queueName string) error {
	_, err := r.Channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (r *RabbitMQContainer) PublishMessage(
	ctx context.Context,
	queueName string,
	body []byte,
) error {
	return r.Channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQContainer) PublishMessageWithHeaders(
	ctx context.Context,
	queue string,
	body []byte,
	headers amqp091.Table,
) error {
	return r.Channel.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		amqp091.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		},
	)
}

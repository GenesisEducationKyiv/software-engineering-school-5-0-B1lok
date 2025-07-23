package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"notification/internal/application/event"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	channel *amqp091.Channel
}

func NewConsumer(channel *amqp091.Channel) *Consumer {
	return &Consumer{
		channel: channel,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler event.Handler) error {
	msgs, err := c.channel.ConsumeWithContext(
		ctx,
		string(handler.GetName()),
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to register consumer for queue %s: %c", handler.GetName(), err,
		)
	}

	go consumeLoop(
		ctx,
		msgs,
		string(handler.GetName()),
		func(ctx context.Context, msg amqp091.Delivery,
		) error {
			if err := handler.Handle(ctx, msg.Body); err != nil {
				log.Error().Err(err).Msg("failed to handle message")
				return nack(msg)
			}
			return ack(msg)
		})

	return nil
}

package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/rabbitmq/amqp091-go"
)

func consumeLoop(
	ctx context.Context,
	msgs <-chan amqp091.Delivery,
	queueName string,
	handler func(context.Context, amqp091.Delivery,
	) error) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Str("queue", queueName).Msg("Stopping consumer")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Info().Str("queue", queueName).Msg("Channel closed, stopping consumer")
				return
			}

			if err := handler(ctx, msg); err != nil {
				log.Error().Str("queue", queueName).Err(err).Msg("failed to process message")
			}
		}
	}
}

func ack(msg amqp091.Delivery) error {
	if err := msg.Ack(false); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}
	return nil
}

func nack(msg amqp091.Delivery) error {
	if err := msg.Nack(false, false); err != nil {
		log.Error().Err(err).Msg("failed to nack message")
	}
	return errors.New("nack issued")
}

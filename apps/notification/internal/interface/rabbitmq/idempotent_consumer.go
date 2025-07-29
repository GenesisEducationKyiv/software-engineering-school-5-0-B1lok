package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/rabbitmq/amqp091-go"

	"notification/internal/application/event"
	"notification/pkg"
)

type Saver interface {
	SaveMessageID(ctx context.Context, messageId string) error
}

type IdempotentConsumer struct {
	channel   *amqp091.Channel
	txManager pkg.TxManager
	saver     Saver
}

func NewIdempotentConsumer(
	channel *amqp091.Channel,
	txManager pkg.TxManager,
	saver Saver,
) *IdempotentConsumer {
	return &IdempotentConsumer{
		channel:   channel,
		txManager: txManager,
		saver:     saver,
	}
}

func (c *IdempotentConsumer) Consume(ctx context.Context, handler event.Handler) error {
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
			"failed to register consumer for queue %s: %w", handler.GetName(), err,
		)
	}

	go consumeLoop(
		ctx,
		msgs,
		string(handler.GetName()),
		func(ctx context.Context, msg amqp091.Delivery,
		) error {
			return c.txManager.ExecuteTx(ctx, func(txCtx context.Context) error {
				messageId, err := getMessageID(msg)
				if err != nil {
					log.Error().Err(err).Msg("failed to get message_id from headers")
					return nack(msg)
				}

				if err := c.saver.SaveMessageID(txCtx, messageId); err != nil {
					return fmt.Errorf("failed to save message_id: %w", err)
				}

				if err := handler.Handle(ctx, msg.Body); err != nil {
					log.Error().Err(err).Msg("failed to handle message")
					return nack(msg)
				}

				return ack(msg)
			})
		})

	return nil
}

func getMessageID(msg amqp091.Delivery) (string, error) {
	val, ok := msg.Headers["message_id"]
	if !ok {
		return "", errors.New("message_id header not found")
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", errors.New("message_id must be non-empty string")
	}
	return str, nil
}

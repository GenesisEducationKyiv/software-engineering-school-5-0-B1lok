package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
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
			log.Printf("Stopping consumer for queue %s", queueName)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Channel closed, stopping consumer for queue %s", queueName)
				return
			}

			if err := handler(ctx, msg); err != nil {
				log.Printf("failed to process message from queue %s: %v", queueName, err)
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
		log.Printf("failed to nack message: %v", err)
	}
	return errors.New("nack issued")
}

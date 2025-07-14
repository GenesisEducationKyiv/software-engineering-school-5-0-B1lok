package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"notification/internal/rabbitmq"
)

type Sender interface {
	ConfirmationEmail(emailMsg *ConfirmationEmailMessage) error
	DailyUpdate(ctx context.Context, updateMsg *WeatherUpdateMessage) error
	HourlyUpdate(ctx context.Context, updateMsg *WeatherUpdateMessage) error
}

type Worker struct {
	channel *amqp091.Channel
	sender  Sender
}

type MessageHandler func(msg amqp091.Delivery) error

func NewWorker(channel *amqp091.Channel, sender Sender) *Worker {
	return &Worker{
		channel: channel,
		sender:  sender,
	}
}

func (w *Worker) StartConfirmationConsumer() error {
	return w.consume(rabbitmq.ConfirmationQueue, w.handleConfirmationEmail)
}

func (w *Worker) StartHourlyUpdateConsumer() error {
	return w.consume(rabbitmq.WeatherHourlyQueue, w.handleWeatherUpdate(w.sender.HourlyUpdate))
}

func (w *Worker) StartDailyUpdateConsumer() error {
	return w.consume(rabbitmq.WeatherDailyQueue, w.handleWeatherUpdate(w.sender.DailyUpdate))
}

func (w *Worker) handleConfirmationEmail(msg amqp091.Delivery) error {
	var emailMsg ConfirmationEmailMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		return fmt.Errorf("failed to parse confirmation email message: %w", err)
	}

	return w.sender.ConfirmationEmail(&emailMsg)
}

func (w *Worker) handleWeatherUpdate(
	updateFunc func(context.Context, *WeatherUpdateMessage,
	) error) MessageHandler {
	return func(msg amqp091.Delivery) error {
		var updateMsg WeatherUpdateMessage
		if err := json.Unmarshal(msg.Body, &updateMsg); err != nil {
			return fmt.Errorf("failed to parse weather update message: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		return updateFunc(ctx, &updateMsg)
	}
}

func (w *Worker) consume(queueName string, handler MessageHandler) error {
	msgs, err := w.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for queue %s: %w", queueName, err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg); err != nil {
				log.Printf("failed to handle message from queue %s: %v", queueName, err)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("failed to nack message: %v", nackErr)
				}
				continue
			}

			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("failed to ack message: %v", ackErr)
			}
		}
	}()

	return nil
}

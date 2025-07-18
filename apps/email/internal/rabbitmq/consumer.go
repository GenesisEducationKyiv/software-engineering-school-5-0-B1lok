package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type Sender interface {
	Send(templateName, to, subject string, data any) error
}

type Consumer struct {
	sender  Sender
	channel *amqp091.Channel
}

type MessageHandler func(msg amqp091.Delivery) error

func NewConsumer(channel *amqp091.Channel, sender Sender) *Consumer {
	return &Consumer{
		sender:  sender,
		channel: channel,
	}
}

func (c *Consumer) StartConfirmationConsumer(ctx context.Context) error {
	return c.consume(ctx, emailConfirmationQueue, c.handleConfirmationEmail)
}

func (c *Consumer) StartHourlyUpdateConsumer(ctx context.Context) error {
	return c.consume(ctx, emailWeatherHourlyQueue, c.handleHourlyUpdate)
}

func (c *Consumer) StartDailyUpdateConsumer(ctx context.Context) error {
	return c.consume(ctx, emailWeatherDailyQueue, c.handleDailyUpdate)
}

func (c *Consumer) handleConfirmationEmail(msg amqp091.Delivery) error {
	var emailMsg ConfirmationEmailMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		return fmt.Errorf("failed to parse confirmation email message: %w", err)
	}

	fmt.Printf("email confirmation email message: %+v\n", emailMsg)
	return c.sender.Send(
		emailMsg.TemplateName,
		emailMsg.To,
		"Confirm your subscription",
		emailMsg.TemplateData,
	)
}

func (c *Consumer) handleHourlyUpdate(msg amqp091.Delivery) error {
	var emailMsg HourlyUpdateMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		return fmt.Errorf("failed to parse hourly update message: %w", err)
	}

	return c.sender.Send(
		emailMsg.TemplateName,
		emailMsg.To,
		"Your weather hourly forecast",
		emailMsg.TemplateData,
	)
}

func (c *Consumer) handleDailyUpdate(msg amqp091.Delivery) error {
	var emailMsg DailyUpdateMessage
	if err := json.Unmarshal(msg.Body, &emailMsg); err != nil {
		return fmt.Errorf("failed to parse daily update message: %w", err)
	}

	return c.sender.Send(
		emailMsg.TemplateName,
		emailMsg.To,
		"Your weather daily forecast",
		emailMsg.TemplateData,
	)
}

func (c *Consumer) consume(ctx context.Context, queueName string, handler MessageHandler) error {
	msgs, err := c.channel.Consume(
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

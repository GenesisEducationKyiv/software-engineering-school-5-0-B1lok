package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const bufferSize = 1000

type publishRequest struct {
	ctx     context.Context
	queue   string
	payload []byte
	done    chan error
}

type Publisher struct {
	channel     *amqp091.Channel
	confirmChan <-chan amqp091.Confirmation
	requests    chan publishRequest
}

func NewPublisher(channel *amqp091.Channel) (*Publisher, error) {
	if err := channel.Confirm(false); err != nil {
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}
	confirmChan := channel.NotifyPublish(make(chan amqp091.Confirmation, 1))

	p := &Publisher{
		channel:     channel,
		confirmChan: confirmChan,
		requests:    make(chan publishRequest, bufferSize),
	}

	go p.startPublisherLoop()

	return p, nil
}

func (p *Publisher) startPublisherLoop() {
	for req := range p.requests {
		err := p.publishWithConfirm(req.ctx, req.queue, req.payload)
		req.done <- err
	}
}

func (p *Publisher) publishWithConfirm(ctx context.Context, queue string, payload []byte) error {
	err := p.channel.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	select {
	case confirm := <-p.confirmChan:
		if !confirm.Ack {
			return fmt.Errorf("message not acknowledged by the broker")
		}
	case <-ctx.Done():
		return fmt.Errorf("publish canceled: %w", ctx.Err())
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timeout waiting for publisher confirmation")
	}

	return nil
}

func (p *Publisher) Publish(ctx context.Context, queue string, payload []byte) error {
	done := make(chan error, 1)

	select {
	case p.requests <- publishRequest{
		ctx:     ctx,
		queue:   queue,
		payload: payload,
		done:    done,
	}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

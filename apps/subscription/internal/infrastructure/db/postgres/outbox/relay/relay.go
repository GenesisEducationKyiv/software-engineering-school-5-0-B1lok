package relay

import (
	"context"
	"fmt"

	"subscription-service/internal/infrastructure/db/postgres/outbox"
	"subscription-service/pkg/middleware"
)

const batchSize = 10

type Publisher interface {
	PublishWithHeaders(
		ctx context.Context,
		queue string,
		payload []byte,
		headers map[string]interface{},
	) error
}

type OutboxRepository interface {
	GetPendingMessages(ctx context.Context, limit int) ([]outbox.Message, error)
	UpdateStatus(ctx context.Context, messageID uint, status outbox.Status) error
}

type Relay struct {
	publisher  Publisher
	txManager  middleware.TxManager
	repository OutboxRepository
}

func NewRelayJob(
	publisher Publisher,
	txManager middleware.TxManager,
	repository OutboxRepository,
) *Relay {
	return &Relay{
		publisher:  publisher,
		txManager:  txManager,
		repository: repository,
	}
}

func (r Relay) Name() string {
	return "Outbox Table Relay"
}

func (r Relay) Schedule() string {
	return "* * * * * *"
}

func (r Relay) Run(ctx context.Context) error {
	return r.txManager.ExecuteTx(ctx, func(txCtx context.Context) error {
		messages, err := r.repository.GetPendingMessages(txCtx, batchSize)
		if err != nil {
			return fmt.Errorf("failed to get pending messages: %w", err)
		}

		if len(messages) == 0 {
			return nil
		}
		for _, msg := range messages {
			headers := map[string]interface{}{
				"message_id": msg.MessageID.String(),
			}
			err := r.publisher.PublishWithHeaders(
				txCtx,
				string(msg.EventType),
				msg.Payload,
				headers,
			)
			newStatus := outbox.StatusSuccess
			if err != nil {
				newStatus = outbox.StatusFailed
			}

			if err := r.repository.UpdateStatus(txCtx, msg.ID, newStatus); err != nil {
				return fmt.Errorf("failed to update outbox status: %w", err)
			}
		}

		return nil
	})
}

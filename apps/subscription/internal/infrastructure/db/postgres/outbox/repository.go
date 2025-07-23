package outbox

import (
	"context"

	"gorm.io/gorm"

	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"
	"subscription-service/pkg/middleware"
)

type Repository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, outbox Message) error {
	db := r.getDB(ctx)

	if err := db.Create(&outbox).Error; err != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to create outbox message",
		)
	}

	return nil
}

func (r *Repository) GetPendingMessages(ctx context.Context, limit int) ([]Message, error) {
	db := r.getDB(ctx)

	var messages []Message

	err := db.Raw(`
			SELECT id, aggregate_id, message_id, event_type, payload, status, created_at, updated_at 
			FROM outbox 
			WHERE status IN (?, ?) 
			ORDER BY created_at 
			LIMIT ? 
			FOR UPDATE SKIP LOCKED
		`, StatusPending, StatusFailed, limit).Scan(&messages).Error

	if err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to get pending messages",
		)
	}

	return messages, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, messageID uint, status Status) error {
	db := r.getDB(ctx)

	result := db.Model(&Message{}).Where("id = ?", messageID).Update("status", status)
	if result.Error != nil {
		return pkgErrors.New(
			internalErrors.ErrInternal, "failed to update message status",
		)
	}

	if result.RowsAffected == 0 {
		return pkgErrors.New(
			internalErrors.ErrInternal, "message not found or not updated",
		)
	}

	return nil
}

func (r *Repository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := middleware.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

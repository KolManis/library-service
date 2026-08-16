package ports

import (
	"context"

	"github.com/google/uuid"
)

// OutboxRecord — запись очереди outbox, которую читает воркер.
type OutboxRecord struct {
	ID        uuid.UUID
	EventType string
	Payload   []byte
}

// IOutboxReader — порт для чтения неопубликованных событий воркером cmd/outbox.
type IOutboxReader interface {
	// FetchUnpublished возвращает до limit записей с published_at IS NULL.
	FetchUnpublished(ctx context.Context, limit int) ([]OutboxRecord, error)
	// MarkPublished помечает запись как опубликованную.
	MarkPublished(ctx context.Context, id uuid.UUID) error
}

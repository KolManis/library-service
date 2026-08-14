package fine

import (
	"time"

	"github.com/KolManis/library-service/internal/core/domain/events"
	"github.com/google/uuid"
)

// FineCreated — событие, которое поднимается при создании штрафа.
type FineCreated struct {
	FineID     uuid.UUID
	LoanID     uuid.UUID
	Amount     float64
	OccurredAt time.Time
}

// EventType возвращает тип события — "fine.created".
func (FineCreated) EventType() string { return "fine.created" }

var _ events.DomainEvent = (*FineCreated)(nil)

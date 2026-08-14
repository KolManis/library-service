package loan

import (
	"time"

	"github.com/KolManis/library-service/internal/core/domain/events"
	"github.com/google/uuid"
)

// LoanChanged — событие смены статуса выдачи.
type LoanChanged struct {
	LoanID     uuid.UUID
	Status     Status
	OccurredAt time.Time
}

// EventType возвращает тип события — "loan.changed".
func (LoanChanged) EventType() string { return "loan.changed" }

var _ events.DomainEvent = (*LoanChanged)(nil)

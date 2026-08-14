package fine

import (
	"math"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/events"
	"github.com/google/uuid"
)

// Status — статус штрафа.
type Status string

const (
	// StatusPending — штраф начислен, но не оплачен.
	StatusPending Status = "pending"
	// StatusPaid — штраф оплачен.
	StatusPaid Status = "paid"
)

// FineRatePerDay — ставка штрафа за один день просрочки.
const FineRatePerDay = 50.0

// Fine — агрегат штрафа за просрочку возврата книги.
type Fine struct {
	id     uuid.UUID
	loanID uuid.UUID
	amount float64
	status Status
	events []events.DomainEvent
}

// NewFine создаёт штраф за просрочку возврата книги. amount считается как
// FineRatePerDay * количество дней между dueAt и returnedAt (округление вверх).
// Возвращает ErrNotOverdue, если returnedAt не позже dueAt.
func NewFine(loanID uuid.UUID, dueAt, returnedAt time.Time) (*Fine, error) {
	if loanID == uuid.Nil {
		return nil, ErrEmptyID
	}

	if !returnedAt.After(dueAt) {
		return nil, ErrNotOverdue
	}

	// Количество полных дней просрочки с округлением вверх
	days := math.Ceil(returnedAt.Sub(dueAt).Hours() / 24)
	amount := FineRatePerDay * days

	f := &Fine{
		id:     uuid.New(),
		loanID: loanID,
		amount: amount,
		status: StatusPending,
	}
	f.events = append(f.events, FineCreated{
		FineID:     f.id,
		LoanID:     f.loanID,
		Amount:     f.amount,
		OccurredAt: returnedAt,
	})
	return f, nil
}

// Restore восстанавливает Fine из хранилища. Без валидации.
func Restore(id, loanID uuid.UUID, amount float64, status Status) *Fine {
	return &Fine{
		id:     id,
		loanID: loanID,
		amount: amount,
		status: status,
	}
}

// ID возвращает идентификатор штрафа.
func (f *Fine) ID() uuid.UUID { return f.id }

// LoanID возвращает идентификатор выдачи, за которую начислен штраф.
func (f *Fine) LoanID() uuid.UUID { return f.loanID }

// Amount возвращает сумму штрафа.
func (f *Fine) Amount() float64 { return f.amount }

// Status возвращает статус штрафа (pending/paid).
func (f *Fine) Status() Status { return f.status }

// PullEvents возвращает накопленные доменные события и очищает хранилище.
// Вызывается application-слоем после сохранения агрегата, чтобы сохранить события в outbox.
func (f *Fine) PullEvents() []events.DomainEvent {
	evs := f.events
	f.events = nil
	return evs
}

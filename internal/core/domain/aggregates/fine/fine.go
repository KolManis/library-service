package fine

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
)

const FineRatePerDay = 50.0

type Fine struct {
	id     uuid.UUID
	loanID uuid.UUID
	amount float64
	status Status
}

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

	return &Fine{
		id:     uuid.New(),
		loanID: loanID,
		amount: amount,
		status: StatusPending,
	}, nil
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

func (f *Fine) ID() uuid.UUID     { return f.id }
func (f *Fine) LoanID() uuid.UUID { return f.loanID }
func (f *Fine) Amount() float64   { return f.amount }
func (f *Fine) Status() Status    { return f.status }

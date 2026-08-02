package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LoanView — плоская read-модель для отображения выдачи.
// Отличается от доменного loan.Loan тем, что не содержит бизнес-правил.
type LoanView struct {
	LoanID     uuid.UUID
	CopyID     uuid.UUID
	Status     string
	ReservedAt time.Time
	IssuedAt   *time.Time
	DueAt      *time.Time
	ReturnedAt *time.Time
}

// IReaderLoansReader — порт для получения списка выдач читателя.
type IReaderLoansReader interface {
	// FindByReaderID возвращает все выдачи (в том числе завершённые) для указанного readerID.
	// Если выдач нет, возвращается пустой срез (nil или []LoanView{}), без ошибки.
	FindByReaderID(ctx context.Context, readerID uuid.UUID) ([]LoanView, error)
}

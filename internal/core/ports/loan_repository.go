package ports

import (
	"context"
	"errors"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/google/uuid"
)

var ErrLoanNotFound = errors.New("выдача не найдена")

// ILoanRepository — порт для хранения агрегата Loan.
// Интерфейс объявлен в ядре, реализуется адаптером — инверсия зависимостей.
type ILoanRepository interface {
	Create(ctx context.Context, l *loan.Loan) error
	GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error)
	Update(ctx context.Context, l *loan.Loan) error
}

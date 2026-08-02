package ports

import (
	"context"
	"errors"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/google/uuid"
)

// ErrLoanNotFound — выдача с указанным id не найдена.
var ErrLoanNotFound = errors.New("выдача не найдена")

// ILoanRepository — порт для хранения агрегата Loan.
// Интерфейс объявлен в ядре, реализуется адаптером — инверсия зависимостей.
type ILoanRepository interface {
	// Create сохраняет новую выдачу.
	Create(ctx context.Context, l *loan.Loan) error
	// GetByID возвращает выдачу по id. Не найдена — ErrLoanNotFound.
	GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error)
	// Update сохраняет изменённое состояние выдачи.
	Update(ctx context.Context, l *loan.Loan) error
}

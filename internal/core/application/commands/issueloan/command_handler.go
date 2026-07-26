package issueloan

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/ports"
)

// Handler выполняет сценарий выдачи книги читателю на руки
type Handler struct {
	loans ports.ILoanRepository
	now   func() time.Time
}

// NewHandler создаёт обработчик, принимая зависимости через интерфейсы
func NewHandler(loans ports.ILoanRepository, now func() time.Time) *Handler {
	return &Handler{
		loans: loans,
		now:   now,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	l, err := h.loans.GetByID(ctx, cmd.LoanID)
	if err != nil {
		return fmt.Errorf("получение выдачи: %w", err)
	}

	if err := l.Issue(h.now()); err != nil {
		return fmt.Errorf("выдача книги: %w", err)
	}

	if err := h.loans.Update(ctx, l); err != nil {
		return fmt.Errorf("сохранение выдачи: %w", err)
	}

	return nil
}

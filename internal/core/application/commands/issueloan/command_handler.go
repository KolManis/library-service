package issueloan

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/ports"
)

// Handler выполняет сценарий выдачи книги читателю на руки.
type Handler struct {
	loans  ports.ILoanRepository
	outbox ports.IOutboxRepository
	tx     ports.ITransactor
	now    func() time.Time
}

// NewHandler создаёт Handler с зависимостями, переданными через порты.
func NewHandler(
	loans ports.ILoanRepository,
	outbox ports.IOutboxRepository,
	tx ports.ITransactor,
	now func() time.Time,
) *Handler {
	return &Handler{
		loans:  loans,
		outbox: outbox,
		tx:     tx,
		now:    now,
	}
}

// Handle переводит бронь в статус issued и выставляет due_at. Ошибки:
// ports.ErrLoanNotFound (нет такой выдачи), loan.ErrInvalidTransition
// (недопустимый переход из текущего статуса).
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
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

		if err := h.outbox.Append(ctx, l.PullEvents()); err != nil {
			return fmt.Errorf("сохранение события: %w", err)
		}
		return nil
	})
}

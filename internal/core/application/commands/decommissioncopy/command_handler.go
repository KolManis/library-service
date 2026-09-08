package decommissioncopy

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/ports"
)

// Handler выполняет сценарий списания одной копии.
type Handler struct {
	copies ports.ICopyRepository
	loans  ports.ILoanRepository
	outbox ports.IOutboxRepository
	tx     ports.ITransactor
	now    func() time.Time
}

// NewHandler создаёт обработчик, принимая зависимости через интерфейсы.
func NewHandler(
	copies ports.ICopyRepository,
	loans ports.ILoanRepository,
	outbox ports.IOutboxRepository,
	tx ports.ITransactor,
	now func() time.Time,
) *Handler {
	return &Handler{
		copies: copies,
		loans:  loans,
		outbox: outbox,
		tx:     tx,
		now:    now,
	}
}

// Handle выполняет списание копии: помечает её decommissioned,
// отменяет связанную активную бронь (если есть), сохраняет изменения
// и пишет события в outbox в одной транзакции.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.copies.DecommissionByID(ctx, cmd.CopyID); err != nil {
			return fmt.Errorf("списание копии: %w", err)
		}

		l, err := h.loans.FindReservedByCopyID(ctx, cmd.CopyID)
		if err != nil {
			return fmt.Errorf("поиск брони: %w", err)
		}

		if l != nil {
			if err := l.Cancel(h.now()); err != nil {
				return fmt.Errorf("отмена брони: %w", err)
			}

			if err := h.loans.Update(ctx, l); err != nil {
				return fmt.Errorf("сохранение брони: %w", err)
			}

			if err := h.outbox.Append(ctx, l.PullEvents()); err != nil {
				return fmt.Errorf("сохранение события: %w", err)
			}
		}

		return nil
	})
}

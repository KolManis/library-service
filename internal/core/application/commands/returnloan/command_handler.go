package returnloan

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

// Handler выполняет сценарий возврата книги.
type Handler struct {
	loans  ports.ILoanRepository
	fines  ports.IFineRepository
	outbox ports.IOutboxRepository
	tx     ports.ITransactor
	now    func() time.Time
}

// NewHandler создаёт Handler с зависимостями, переданными через порты.
func NewHandler(
	loans ports.ILoanRepository,
	fines ports.IFineRepository,
	outbox ports.IOutboxRepository,
	tx ports.ITransactor,
	now func() time.Time,
) *Handler {
	return &Handler{
		loans:  loans,
		fines:  fines,
		outbox: outbox,
		tx:     tx,
		now:    now,
	}
}

// Handle выполняет возврат книги. Если книга была просрочена, создаёт штраф.
// Все операции выполняются в одной транзакции.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		// 1. Получаем выдачу
		l, err := h.loans.GetByID(ctx, cmd.LoanID)
		if err != nil {
			return fmt.Errorf("получение выдачи: %w", err)
		}

		// 2. Запоминаем, была ли просрочка до вызова Return,
		//    потому что после Return статус изменится.
		wasOverdue := l.Status() == loan.StatusOverdue
		now := h.now()

		// 3. Выполняем возврат (домен проверяет допустимость перехода)
		if err := l.Return(now); err != nil {
			return fmt.Errorf("возврат книги: %w", err)
		}

		// 4. Сохраняем изменённый Loan
		if err := h.loans.Update(ctx, l); err != nil {
			return fmt.Errorf("сохранение выдачи: %w", err)
		}

		events := l.PullEvents()
		// 5. Если была просрочка — создаём штраф.
		//    DueAt не меняется при Return, можно читать после. Разыменовываем
		//    без nil-проверки: overdue достижим только через MarkOverdue из
		//    issued, а тот всегда ставит dueAt — тот же инвариант, на который
		//    полагается сам MarkOverdue (см. loan.go).
		if wasOverdue {
			dueAt := l.DueAt()
			if dueAt == nil {
				return fmt.Errorf("dueAt отсутствует у просроченной выдачи")
			}
			f, err := fine.NewFine(l.ID(), *dueAt, now)
			if err != nil {
				return fmt.Errorf("создание штрафа: %w", err)
			}
			if err := h.fines.Create(ctx, f); err != nil {
				return fmt.Errorf("сохранение штрафа: %w", err)
			}
			events = append(events, f.PullEvents()...)
		}

		if err := h.outbox.Append(ctx, events); err != nil {
			return fmt.Errorf("сохранение событий: %w", err)
		}
		return nil
	})
}

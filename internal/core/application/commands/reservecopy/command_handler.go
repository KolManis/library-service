package reservecopy

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
)

// Handler выполняет сценарий бронирования экземпляра книги.
type Handler struct {
	copies ports.ICopyRepository
	loans  ports.ILoanRepository
	outbox ports.IOutboxRepository
	tx     ports.ITransactor
	now    func() time.Time
}

// NewHandler создаёт Handler с зависимостями, переданными через порты.
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

// Handle бронирует свободный экземпляр книги за читателем. Возвращает id
// созданной брони или ошибку (ports.ErrNoFreeCopy, если свободных нет).
func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	var loanID uuid.UUID

	err := h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		// 1. найти свободный экземпляр книги
		copyID, err := h.copies.FindFreeCopyID(ctx, cmd.BookID)
		if err != nil {
			return fmt.Errorf("поиск свободного экземпляра: %w", err)
		}

		// 2. создать бронь через доменный конструктор (там бизнес-правила)
		l, err := loan.Reserve(copyID, cmd.ReaderID, h.now())
		if err != nil {
			return fmt.Errorf("создание брони: %w", err)
		}

		// 3. сохранить бронь в хранилище
		if err := h.loans.Create(ctx, l); err != nil {
			return fmt.Errorf("сохранение брони: %w", err)
		}

		if err := h.outbox.Append(ctx, l.PullEvents()); err != nil {
			return fmt.Errorf("сохранение события: %w", err)
		}
		loanID = l.ID()
		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return loanID, nil
}

package reservecopy

import (
	"context"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
)

// Handler выполняет сценарий бронирования экземпляра книги
type Handler struct {
	copies ports.ICopyRepository // порт для поиска свободного экземпляра
	loans  ports.ILoanRepository // порт для сохранения выдачи
	now    func() time.Time      // функция-источник времени (в проде time.Now, в тесте фиксированная)
}

// NewHandler создаёт обработчик, принимая зависимости через интерфейсы
func NewHandler(
	copies ports.ICopyRepository,
	loans ports.ILoanRepository,
	now func() time.Time,
) *Handler {
	return &Handler{
		copies: copies,
		loans:  loans,
		now:    now,
	}
}

// Handle запускает сценарий бронирования.
// Возвращает ID созданной выдачи (loanID) или ошибку.
func (h *Handler) Handle(ctx context.Context, cmd Command) (uuid.UUID, error) {
	// Шаг 1: найти свободный экземпляр книги
	copyID, err := h.copies.FindFreeCopyID(ctx, cmd.BookID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("поиск свободного экземпляра: %w", err)
	}

	// Шаг 2: создать бронь через доменный конструктор (там бизнес-правила)
	l, err := loan.Reserve(copyID, cmd.ReaderID, h.now())
	if err != nil {
		return uuid.Nil, fmt.Errorf("создание брони: %w", err)
	}

	// Шаг 3: сохранить бронь в хранилище
	if err := h.loans.Create(ctx, l); err != nil {
		return uuid.Nil, fmt.Errorf("сохранение брони: %w", err)
	}

	return l.ID(), nil
}

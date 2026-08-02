package getreaderloans

import (
	"context"

	"github.com/KolManis/library-service/internal/core/ports"
)

// Handler выполняет запрос списка выдач читателя.
type Handler struct {
	reader ports.IReaderLoansReader
}

// NewHandler создаёт Handler с зависимостью, переданной через порт.
func NewHandler(reader ports.IReaderLoansReader) *Handler {
	return &Handler{
		reader: reader,
	}
}

// Handle возвращает все выдачи читателя (включая завершённые). Пустой список
// — не ошибка.
func (h *Handler) Handle(ctx context.Context, q Query) ([]ports.LoanView, error) {
	return h.reader.FindByReaderID(ctx, q.ReaderID)
}

package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNoFreeCopy = errors.New("нет свободных экземпляров")

// ICopyRepository - порт для работы с экземплярами книг.
type ICopyRepository interface {
	// FindFreeCopyID возвращает id свободного экземпляра книги:
	// не списанного и без активной выдачи. Нет такого - ErrNoFreeCopy.
	FindFreeCopyID(ctx context.Context, bookID uuid.UUID) (uuid.UUID, error)
}

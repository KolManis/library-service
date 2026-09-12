package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrNoFreeCopy — среди экземпляров книги нет ни одного свободного.
var ErrNoFreeCopy = errors.New("нет свободных экземпляров")

// ErrCopyNotFound — копия с указанным ID не найдена.
var ErrCopyNotFound = errors.New("копия не найдена")

// ICopyRepository - порт для работы с экземплярами книг.
type ICopyRepository interface {
	// FindFreeCopyID возвращает id свободного экземпляра книги:
	// не списанного и без активной выдачи. Нет такого - ErrNoFreeCopy.
	FindFreeCopyID(ctx context.Context, bookID uuid.UUID) (uuid.UUID, error)
	// DecommissionByID помечает копию как списанную по её ID.
	// Если копия не найдена — возвращает ErrCopyNotFound.
	DecommissionByID(ctx context.Context, copyID uuid.UUID) error
}

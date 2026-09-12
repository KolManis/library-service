package decommissioncopy

import (
	"errors"

	"github.com/google/uuid"
)

// ErrEmptyID — ошибка валидации, если CopyID пустой.
var ErrEmptyID = errors.New("идентификатор не может быть пустым")

// Command - данные, необходимые для списания экземпляра книги
type Command struct {
	CopyID uuid.UUID
}

// NewCommand создаёт команду, проверяя, что CopyID задан
func NewCommand(copyID uuid.UUID) (Command, error) {
	if copyID == uuid.Nil {
		return Command{}, ErrEmptyID
	}
	return Command{CopyID: copyID}, nil
}

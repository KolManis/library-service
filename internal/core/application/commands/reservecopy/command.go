package reservecopy

import (
	"errors"

	"github.com/google/uuid"
)

// ErrEmptyID - ошибка валидации, если BookID или ReaderID пустые
var ErrEmptyID = errors.New("идентификатор не может быть пустым")

// Command - данные, необходимые для бронирования экземпляра книги
type Command struct {
	BookID   uuid.UUID
	ReaderID uuid.UUID
}

// NewCommand создаёт команду, проверяя, что оба идентификатора заданы
func NewCommand(bookID, readerID uuid.UUID) (Command, error) {
	if bookID == uuid.Nil || readerID == uuid.Nil {
		return Command{}, ErrEmptyID
	}
	return Command{BookID: bookID, ReaderID: readerID}, nil
}

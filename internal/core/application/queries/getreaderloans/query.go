package getreaderloans

import (
	"errors"

	"github.com/google/uuid"
)

// ErrEmptyID — в запрос передан нулевой ReaderID.
var ErrEmptyID = errors.New("идентификатор не может быть пустым")

// Query — данные, необходимые для получения списка выдач читателя.
type Query struct {
	ReaderID uuid.UUID
}

// NewQuery создаёт запрос, проверяя, что ReaderID задан.
func NewQuery(readerID uuid.UUID) (Query, error) {
	if readerID == uuid.Nil {
		return Query{}, ErrEmptyID
	}
	return Query{
		ReaderID: readerID,
	}, nil
}

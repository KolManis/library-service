package returnloan

import (
	"errors"

	"github.com/google/uuid"
)

// ErrEmptyID — в команду передан нулевой LoanID.
var ErrEmptyID = errors.New("идентификатор не может быть пустым")

// Command — данные, необходимые для возврата книги.
type Command struct {
	LoanID uuid.UUID
}

// NewCommand создаёт команду, проверяя, что LoanID задан.
func NewCommand(loanID uuid.UUID) (Command, error) {
	if loanID == uuid.Nil {
		return Command{}, ErrEmptyID
	}
	return Command{LoanID: loanID}, nil
}

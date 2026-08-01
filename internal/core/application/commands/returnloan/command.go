package returnloan

import (
	"errors"

	"github.com/google/uuid"
)

var ErrEmptyID = errors.New("идентификатор не может быть пустым")

type Command struct {
	LoanID uuid.UUID
}

func NewCommand(loanID uuid.UUID) (Command, error) {
	if loanID == uuid.Nil {
		return Command{}, ErrEmptyID
	}
	return Command{LoanID: loanID}, nil
}

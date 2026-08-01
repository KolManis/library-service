package fine

import "errors"

var (
	ErrEmptyID    = errors.New("идентификатор не может быть пустым")
	ErrNotOverdue = errors.New("книга не просрочена, штраф не начисляется")
)

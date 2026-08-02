package fine

import "errors"

var (
	// ErrEmptyID — в конструктор передан нулевой uuid.
	ErrEmptyID = errors.New("идентификатор не может быть пустым")
	// ErrNotOverdue — попытка создать штраф без фактической просрочки (returnedAt не позже dueAt).
	ErrNotOverdue = errors.New("книга не просрочена, штраф не начисляется")
)

package loan

import "errors"

var (
	// ErrInvalidTransition — попытка перехода, запрещённого машиной статусов.
	ErrInvalidTransition = errors.New("недопустимый переход статуса")
	// ErrNotYetDue — действие раньше срока: expire до истечения TTL,
	// overdue до наступления due_at.
	ErrNotYetDue = errors.New("срок ещё не наступил")
	// ErrEmptyID — в конструктор передан нулевой uuid.
	ErrEmptyID = errors.New("пустой идентификатор")
)

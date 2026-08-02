package ports

import "context"

// ITransactor обеспечивает атомарное выполнение группы операций.
// Реализация (в адаптере) открывает транзакцию БД и передаёт её
// через контекст внутрь fn, чтобы репозитории могли к ней подключиться.
type ITransactor interface {
	// WithinTransaction выполняет fn в рамках одной транзакции: ошибка fn —
	// откат, nil — коммит.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

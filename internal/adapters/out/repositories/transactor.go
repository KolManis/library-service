package repositories

import (
	"context"

	"gorm.io/gorm"
)

// txKey — ключ для передачи транзакции через контекст.
type txKey struct{}

// Transactor реализует ports.ITransactor.
type Transactor struct {
	db *gorm.DB
}

// NewTransactor создаёт Transactor поверх открытого соединения gorm.
func NewTransactor(db *gorm.DB) *Transactor {
	return &Transactor{db: db}
}

// WithinTransaction открывает транзакцию и выполняет fn.
// Если fn возвращает ошибку, транзакция откатывается.
func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Прокидываем tx через контекст, чтобы репозитории могли его использовать.
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFromContext возвращает *gorm.DB из контекста, если там есть транзакция.
// Иначе использует fallback-соединение.
func dbFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback.WithContext(ctx)
}

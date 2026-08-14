package ports

import (
	"context"

	"github.com/KolManis/library-service/internal/core/domain/events"
)

// IOutboxRepository — порт для надёжной записи доменных событий.
type IOutboxRepository interface {
	// Append сохраняет события. Вызывается внутри той же транзакции,
	// что и Update/Create агрегата, породившего события.
	Append(ctx context.Context, evs []events.DomainEvent) error
}

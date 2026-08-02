package ports

import (
	"context"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
)

// IFineRepository — порт для хранения агрегата Fine.
type IFineRepository interface {
	// Create сохраняет новый штраф.
	Create(ctx context.Context, f *fine.Fine) error
}

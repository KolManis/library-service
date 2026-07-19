package repositories

import (
	"context"
	"fmt"

	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CopyRepository struct {
	db *gorm.DB
}

func NewCopyRepository(db *gorm.DB) *CopyRepository {
	return &CopyRepository{db: db}
}

func (r *CopyRepository) FindFreeCopyID(ctx context.Context, bookID uuid.UUID) (uuid.UUID, error) {
	// сканируем в структуру, а не в голый uuid.UUID: для полей структуры
	// gorm использует конвертер uuid, для одиночной переменной-массива — нет
	var row struct{ ID uuid.UUID }
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id FROM copies c
		WHERE c.book_id = ?
			AND NOT c.decommissioned
			AND NOT EXISTS (
		      SELECT 1 FROM loans l
		      WHERE l.copy_id = c.id AND l.status IN ('reserved', 'issued')
		  )
		LIMIT 1`, bookID).Scan(&row).Error
	if err != nil {
		return uuid.Nil, err
	}
	if row.ID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: bookID=%s", ports.ErrNoFreeCopy, bookID)
	}
	return row.ID, nil
}

package repositories

import (
	"context"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fineModel — строка таблицы fines. Знает про БД всё,
// про бизнес-правила — ничего.
type fineModel struct {
	ID     uuid.UUID `gorm:"primaryKey"`
	LoanID uuid.UUID
	Amount float64
	Status string
}

// TableName говорит gorm имя таблицы — иначе он выведет "fine_models".
func (fineModel) TableName() string { return "fines" }

func toFineModel(f *fine.Fine) fineModel {
	return fineModel{
		ID:     f.ID(),
		LoanID: f.LoanID(),
		Amount: f.Amount(),
		Status: string(f.Status()),
	}
}

// FineRepository — gorm-реализация ports.IFineRepository.
type FineRepository struct {
	db *gorm.DB
}

// NewFineRepository создаёт FineRepository поверх открытого соединения gorm.
func NewFineRepository(db *gorm.DB) *FineRepository {
	return &FineRepository{db: db}
}

// Create сохраняет новый штраф.
func (r *FineRepository) Create(ctx context.Context, f *fine.Fine) error {
	m := toFineModel(f)
	return dbFromContext(ctx, r.db).Create(&m).Error
}

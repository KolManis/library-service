package repositories

import (
	"context"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fineModel struct {
	ID     uuid.UUID `gorm:"primaryKey"`
	LoanID uuid.UUID
	Amount float64
	Status string
}

func (fineModel) TableName() string { return "fines" }

func toFineModel(f *fine.Fine) fineModel {
	return fineModel{
		ID:     f.ID(),
		LoanID: f.LoanID(),
		Amount: f.Amount(),
		Status: string(f.Status()),
	}
}

type FineRepository struct {
	db *gorm.DB
}

func NewFineRepository(db *gorm.DB) *FineRepository {
	return &FineRepository{db: db}
}

func (r *FineRepository) Create(ctx context.Context, f *fine.Fine) error {
	m := toFineModel(f)
	return dbFromContext(ctx, r.db).Create(&m).Error
}

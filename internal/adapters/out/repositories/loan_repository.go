package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// loanModel — строка таблицы loans. Знает про БД всё,
// про бизнес-правила — ничего.
type loanModel struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	CopyID     uuid.UUID
	ReaderID   uuid.UUID
	Status     string
	ReservedAt time.Time
	IssuedAt   *time.Time
	DueAt      *time.Time
	ReturnedAt *time.Time
}

// TableName говорит gorm имя таблицы — иначе он выведет "loan_models".
func (loanModel) TableName() string { return "loans" }

func toModel(l *loan.Loan) loanModel {
	return loanModel{
		ID:         l.ID(),
		CopyID:     l.CopyID(),
		ReaderID:   l.ReaderID(),
		Status:     string(l.Status()),
		ReservedAt: l.ReservedAt(),
		IssuedAt:   l.IssuedAt(),
		DueAt:      l.DueAt(),
		ReturnedAt: l.ReturnedAt(),
	}
}

func toDomain(m loanModel) *loan.Loan {
	return loan.Restore(
		m.ID, m.CopyID, m.ReaderID,
		loan.Status(m.Status),
		m.ReservedAt,
		m.IssuedAt, m.DueAt, m.ReturnedAt,
	)
}

type LoanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, l *loan.Loan) error {
	m := toModel(l)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *LoanRepository) GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error) {
	var m loanModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ports.ErrLoanNotFound, id)
		}
		return nil, err
	}
	return toDomain(m), nil
}

func (r *LoanRepository) Update(ctx context.Context, l *loan.Loan) error {
	m := toModel(l)
	return r.db.WithContext(ctx).Save(&m).Error
}

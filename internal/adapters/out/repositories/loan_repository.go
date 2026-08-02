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

func toLoanModel(l *loan.Loan) loanModel {
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

func toLoanDomain(m loanModel) *loan.Loan {
	return loan.Restore(
		m.ID, m.CopyID, m.ReaderID,
		loan.Status(m.Status),
		m.ReservedAt,
		m.IssuedAt, m.DueAt, m.ReturnedAt,
	)
}

// LoanRepository — gorm-реализация ports.ILoanRepository и ports.IReaderLoansReader.
type LoanRepository struct {
	db *gorm.DB
}

// NewLoanRepository создаёт LoanRepository поверх открытого соединения gorm.
func NewLoanRepository(db *gorm.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// Create сохраняет новую выдачу.
func (r *LoanRepository) Create(ctx context.Context, l *loan.Loan) error {
	m := toLoanModel(l)
	return dbFromContext(ctx, r.db).Create(&m).Error
}

// GetByID возвращает выдачу по id. Не найдена — ports.ErrLoanNotFound.
func (r *LoanRepository) GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error) {
	var m loanModel
	err := dbFromContext(ctx, r.db).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ports.ErrLoanNotFound, id)
		}
		return nil, err
	}
	return toLoanDomain(m), nil
}

// Update сохраняет изменённое состояние выдачи.
func (r *LoanRepository) Update(ctx context.Context, l *loan.Loan) error {
	m := toLoanModel(l)
	return dbFromContext(ctx, r.db).Save(&m).Error
}

// FindByReaderID возвращает все выдачи читателя (включая завершённые). Нет
// выдач — пустой срез, без ошибки.
func (r *LoanRepository) FindByReaderID(ctx context.Context, readerID uuid.UUID) ([]ports.LoanView, error) {
	var models []loanModel
	if err := dbFromContext(ctx, r.db).
		Where("reader_id = ?", readerID).
		Order("reserved_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	views := make([]ports.LoanView, 0, len(models))
	for _, m := range models {
		views = append(views, ports.LoanView{
			LoanID:     m.ID,
			CopyID:     m.CopyID,
			Status:     m.Status,
			ReservedAt: m.ReservedAt,
			IssuedAt:   m.IssuedAt,
			DueAt:      m.DueAt,
			ReturnedAt: m.ReturnedAt,
		})
	}
	return views, nil
}

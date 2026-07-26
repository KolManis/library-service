package issueloan

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

// fakeLoanRepo — ручная заглушка для ports.ILoanRepository.
type fakeLoanRepo struct {
	byID    *loan.Loan // что вернуть из GetByID
	getErr  error      // какую ошибку вернуть из GetByID
	updated *loan.Loan // сюда запишется переданный Loan при Update
	updErr  error      // какую ошибку вернуть из Update
}

func (f *fakeLoanRepo) Create(ctx context.Context, l *loan.Loan) error {
	return nil // в этих тестах не используется
}

func (f *fakeLoanRepo) GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error) {
	return f.byID, f.getErr
}

func (f *fakeLoanRepo) Update(ctx context.Context, l *loan.Loan) error {
	f.updated = l
	return f.updErr
}

func TestHandler_Handle_Success(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	reservedAt := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 26, 9, 0, 0, 0, time.UTC)

	l, err := loan.Reserve(copyID, readerID, reservedAt)
	require.NoError(t, err)

	loans := &fakeLoanRepo{byID: l}
	handler := NewHandler(loans, func() time.Time { return now })

	cmd, err := NewCommand(l.ID())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, loans.updated, "Update должен был вызваться")
	require.Equal(t, loan.StatusIssued, loans.updated.Status())
	require.NotNil(t, loans.updated.IssuedAt())
	require.True(t, loans.updated.IssuedAt().Equal(now))
	require.NotNil(t, loans.updated.DueAt())
	require.True(t, loans.updated.DueAt().Equal(now.AddDate(0, 0, loan.LoanPeriodDays)))
}

func TestHandler_Handle_NotFound(t *testing.T) {
	loans := &fakeLoanRepo{getErr: ports.ErrLoanNotFound}
	handler := NewHandler(loans, time.Now)

	cmd, err := NewCommand(uuid.New())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, ports.ErrLoanNotFound)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
}

func TestHandler_Handle_InvalidTransition(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	reservedAt := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	issuedAt := reservedAt.AddDate(0, 0, 1)
	dueAt := issuedAt.AddDate(0, 0, loan.LoanPeriodDays)

	// Loan уже в статусе issued — Restore не валидирует переход,
	// это осознанный чёрный ход для восстановления из БД (см. loan.go).
	l := loan.Restore(uuid.New(), copyID, readerID, loan.StatusIssued, reservedAt, &issuedAt, &dueAt, nil)

	loans := &fakeLoanRepo{byID: l}
	handler := NewHandler(loans, time.Now)

	cmd, err := NewCommand(l.ID())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, loan.ErrInvalidTransition)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
}

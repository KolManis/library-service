package returnloan

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
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

// fakeFineRepo — ручная заглушка для ports.IFineRepository.
type fakeFineRepo struct {
	created *fine.Fine // сюда запишется переданный Fine при Create
	err     error      // какую ошибку вернуть из Create
}

func (f *fakeFineRepo) Create(ctx context.Context, fn *fine.Fine) error {
	f.created = fn
	return f.err
}

// fakeTransactor — ручная заглушка для ports.ITransactor: без реальной БД
// и без реального отката, просто выполняет fn в том же контексте.
type fakeTransactor struct{}

func (fakeTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestHandler_Handle_ReturnOnTime_NoFine(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	issuedAt := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	dueAt := issuedAt.AddDate(0, 0, loan.LoanPeriodDays)
	now := dueAt.Add(-time.Hour) // вернули за час до срока — просрочки не было

	l := loan.Restore(uuid.New(), copyID, readerID, loan.StatusIssued, issuedAt, &issuedAt, &dueAt, nil)

	loans := &fakeLoanRepo{byID: l}
	fines := &fakeFineRepo{}
	handler := NewHandler(loans, fines, fakeTransactor{}, func() time.Time { return now })

	cmd, err := NewCommand(l.ID())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, loans.updated, "Update должен был вызваться")
	require.Equal(t, loan.StatusReturned, loans.updated.Status())
	require.NotNil(t, loans.updated.ReturnedAt())
	require.True(t, loans.updated.ReturnedAt().Equal(now))
	require.Nil(t, fines.created, "штраф не должен был создаться при возврате в срок")
}

func TestHandler_Handle_ReturnOverdue_CreatesFine(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	issuedAt := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	dueAt := issuedAt.AddDate(0, 0, loan.LoanPeriodDays)
	now := dueAt.AddDate(0, 0, 3) // вернули через 3 дня после срока

	l := loan.Restore(uuid.New(), copyID, readerID, loan.StatusOverdue, issuedAt, &issuedAt, &dueAt, nil)

	loans := &fakeLoanRepo{byID: l}
	fines := &fakeFineRepo{}
	handler := NewHandler(loans, fines, fakeTransactor{}, func() time.Time { return now })

	cmd, err := NewCommand(l.ID())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, loans.updated)
	require.Equal(t, loan.StatusReturned, loans.updated.Status())

	require.NotNil(t, fines.created, "штраф должен был создаться при возврате из overdue")
	require.Equal(t, l.ID(), fines.created.LoanID())
	require.Equal(t, fine.StatusPending, fines.created.Status())
	require.Equal(t, fine.FineRatePerDay*3, fines.created.Amount())
}

func TestHandler_Handle_NotFound(t *testing.T) {
	loans := &fakeLoanRepo{getErr: ports.ErrLoanNotFound}
	fines := &fakeFineRepo{}
	handler := NewHandler(loans, fines, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(uuid.New())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, ports.ErrLoanNotFound)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
	require.Nil(t, fines.created, "Create не должен был вызываться")
}

func TestHandler_Handle_InvalidTransition(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	reservedAt := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	// Loan ещё в статусе reserved — Return из reserved недопустим.
	l := loan.Restore(uuid.New(), copyID, readerID, loan.StatusReserved, reservedAt, nil, nil, nil)

	loans := &fakeLoanRepo{byID: l}
	fines := &fakeFineRepo{}
	handler := NewHandler(loans, fines, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(l.ID())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, loan.ErrInvalidTransition)
	require.Nil(t, loans.updated)
	require.Nil(t, fines.created)
}

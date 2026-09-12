package decommissioncopy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/domain/events"
	"github.com/KolManis/library-service/internal/core/ports"
)

// fakeCopyRepo — ручная заглушка для ports.ICopyRepository.
type fakeCopyRepo struct {
	err error // какую ошибку вернуть из DecommissionByID
}

func (f *fakeCopyRepo) FindFreeCopyID(ctx context.Context, bookID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil // в этих тестах не используется
}

func (f *fakeCopyRepo) DecommissionByID(ctx context.Context, copyID uuid.UUID) error {
	return f.err
}

// fakeLoanRepo — ручная заглушка для ports.ILoanRepository.
type fakeLoanRepo struct {
	found   *loan.Loan // что вернуть из FindReservedByCopyID
	findErr error      // какую ошибку вернуть из FindReservedByCopyID
	updated *loan.Loan // сюда запишется переданный Loan при Update
	updErr  error      // какую ошибку вернуть из Update
}

func (f *fakeLoanRepo) Create(ctx context.Context, l *loan.Loan) error {
	return nil // в этих тестах не используется
}

func (f *fakeLoanRepo) GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error) {
	return nil, nil // в этих тестах не используется
}

func (f *fakeLoanRepo) Update(ctx context.Context, l *loan.Loan) error {
	f.updated = l
	return f.updErr
}

func (f *fakeLoanRepo) FindReservedByCopyID(ctx context.Context, copyID uuid.UUID) (*loan.Loan, error) {
	return f.found, f.findErr
}

// fakeOutboxRepo — ручная заглушка для ports.IOutboxRepository.
type fakeOutboxRepo struct {
	appended []events.DomainEvent
	err      error
}

func (f *fakeOutboxRepo) Append(ctx context.Context, evs []events.DomainEvent) error {
	f.appended = append(f.appended, evs...)
	return f.err
}

// fakeTransactor — ручная заглушка для ports.ITransactor: без реальной БД
// и без реального отката, просто выполняет fn в том же контексте.
type fakeTransactor struct{}

func (fakeTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// TestHandler_Handle_Success_WithReservedLoan — на копии есть reserved-бронь:
// она отменяется, Update и Append вызываются.
func TestHandler_Handle_Success_WithReservedLoan(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	reservedAt := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 26, 9, 0, 0, 0, time.UTC)

	l, err := loan.Reserve(copyID, readerID, reservedAt)
	require.NoError(t, err)

	loans := &fakeLoanRepo{found: l}
	outbox := &fakeOutboxRepo{}
	handler := NewHandler(&fakeCopyRepo{}, loans, outbox, fakeTransactor{}, func() time.Time { return now })

	cmd, err := NewCommand(copyID)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, loans.updated, "Update должен был вызваться")
	require.Equal(t, loan.StatusCancelled, loans.updated.Status())
	require.NotEmpty(t, outbox.appended, "событие отмены должно было уйти в outbox")
}

// TestHandler_Handle_Success_NoReservedLoan — на копии нет reserved-брони:
// копия списывается, но Update/Append не вызываются — отменять нечего.
func TestHandler_Handle_Success_NoReservedLoan(t *testing.T) {
	loans := &fakeLoanRepo{found: nil}
	outbox := &fakeOutboxRepo{}
	handler := NewHandler(&fakeCopyRepo{}, loans, outbox, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(uuid.New())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Nil(t, loans.updated, "Update не должен был вызываться")
	require.Empty(t, outbox.appended, "нечего было отменять — событий быть не должно")
}

// TestHandler_Handle_CopyNotFound — копия не найдена: ошибка пробрасывается,
// до поиска брони дело не доходит.
func TestHandler_Handle_CopyNotFound(t *testing.T) {
	copies := &fakeCopyRepo{err: ports.ErrCopyNotFound}
	loans := &fakeLoanRepo{}
	handler := NewHandler(copies, loans, &fakeOutboxRepo{}, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(uuid.New())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, ports.ErrCopyNotFound)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
}

// TestHandler_Handle_FindReservedError — сбой при поиске брони пробрасывается
// как есть, копия при этом уже могла быть списана — транзакция это откатит.
func TestHandler_Handle_FindReservedError(t *testing.T) {
	findErr := errors.New("ошибка БД")
	loans := &fakeLoanRepo{findErr: findErr}
	handler := NewHandler(&fakeCopyRepo{}, loans, &fakeOutboxRepo{}, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(uuid.New())
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, findErr)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
}

// TestHandler_Handle_InvalidTransition — защитный кейс: если найденная бронь
// почему-то не в reserved (контракт FindReservedByCopyID нарушен), Cancel
// должен отказать, а не отменить что попало.
func TestHandler_Handle_InvalidTransition(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	reservedAt := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	issuedAt := reservedAt.AddDate(0, 0, 1)
	dueAt := issuedAt.AddDate(0, 0, loan.LoanPeriodDays)

	// Loan уже в статусе issued — Restore не валидирует переход,
	// это осознанный чёрный ход для восстановления из БД (см. loan.go).
	l := loan.Restore(uuid.New(), copyID, readerID, loan.StatusIssued, reservedAt, &issuedAt, &dueAt, nil)

	loans := &fakeLoanRepo{found: l}
	outbox := &fakeOutboxRepo{}
	handler := NewHandler(&fakeCopyRepo{}, loans, outbox, fakeTransactor{}, time.Now)

	cmd, err := NewCommand(copyID)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, loan.ErrInvalidTransition)
	require.Nil(t, loans.updated, "Update не должен был вызываться")
	require.Empty(t, outbox.appended)
}

package reservecopy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

// fakeCopyRepo — ручная заглушка для ICopyRepository.
type fakeCopyRepo struct {
	id  uuid.UUID // какой ID вернуть
	err error     // какую ошибку вернуть
}

func (f *fakeCopyRepo) FindFreeCopyID(ctx context.Context, bookID uuid.UUID) (uuid.UUID, error) {
	return f.id, f.err
}

// fakeLoanRepo — ручная заглушка для ILoanRepository.
type fakeLoanRepo struct {
	created *loan.Loan // сюда запишем переданный Loan при Create
	err     error      // какую ошибку вернуть из Create
}

func (f *fakeLoanRepo) Create(ctx context.Context, l *loan.Loan) error {
	f.created = l
	return f.err
}

func (f *fakeLoanRepo) GetByID(ctx context.Context, id uuid.UUID) (*loan.Loan, error) {
	return nil, nil // в этих тестах не используется
}

func (f *fakeLoanRepo) Update(ctx context.Context, l *loan.Loan) error {
	return nil // в этих тестах не используется
}

func TestHandler_Handle_Success(t *testing.T) {
	copyID := uuid.New()
	readerID := uuid.New()
	bookID := uuid.New()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)

	copies := &fakeCopyRepo{id: copyID}
	loans := &fakeLoanRepo{}
	handler := NewHandler(copies, loans, func() time.Time { return now })

	cmd, err := NewCommand(bookID, readerID)
	require.NoError(t, err)

	loanID, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotEqual(t, uuid.Nil, loanID)
	require.NotNil(t, loans.created, "Create должен был вызваться")
	require.Equal(t, loan.StatusReserved, loans.created.Status())
	require.Equal(t, copyID, loans.created.CopyID())
	require.Equal(t, readerID, loans.created.ReaderID())
	require.True(t, loans.created.ReservedAt().Equal(now))
}

func TestHandler_Handle_NoFreeCopy(t *testing.T) {
	copies := &fakeCopyRepo{err: ports.ErrNoFreeCopy}
	loans := &fakeLoanRepo{}
	handler := NewHandler(copies, loans, time.Now)

	cmd, err := NewCommand(uuid.New(), uuid.New())
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, ports.ErrNoFreeCopy)
	require.Nil(t, loans.created, "Create не должен был вызываться")
}

func TestHandler_Handle_CreateError(t *testing.T) {
	createErr := errors.New("ошибка БД")
	copies := &fakeCopyRepo{id: uuid.New()}
	loans := &fakeLoanRepo{err: createErr}
	handler := NewHandler(copies, loans, time.Now)

	cmd, err := NewCommand(uuid.New(), uuid.New())
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), cmd)
	require.ErrorIs(t, err, createErr)
}

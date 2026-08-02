package getreaderloans

import (
	"context"
	"errors"
	"testing"

	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeReaderLoansReader — ручная заглушка для ports.IReaderLoansReader.
type fakeReaderLoansReader struct {
	views []ports.LoanView
	err   error
}

func (f *fakeReaderLoansReader) FindByReaderID(ctx context.Context, readerID uuid.UUID) ([]ports.LoanView, error) {
	return f.views, f.err
}

func TestHandler_Handle_Success(t *testing.T) {
	views := []ports.LoanView{
		{LoanID: uuid.New(), Status: "issued"},
		{LoanID: uuid.New(), Status: "returned"},
	}
	reader := &fakeReaderLoansReader{views: views}
	handler := NewHandler(reader)

	q, err := NewQuery(uuid.New())
	require.NoError(t, err)

	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)
	require.Equal(t, views, result)
}

func TestHandler_Handle_Empty(t *testing.T) {
	reader := &fakeReaderLoansReader{views: []ports.LoanView{}} // пустой список — не ошибка
	handler := NewHandler(reader)

	q, err := NewQuery(uuid.New())
	require.NoError(t, err)

	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestHandler_Handle_Error(t *testing.T) {
	dbErr := errors.New("connection lost")
	reader := &fakeReaderLoansReader{err: dbErr}
	handler := NewHandler(reader)

	q, err := NewQuery(uuid.New())
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), q)
	require.ErrorIs(t, err, dbErr)
}

package fine

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewFine(t *testing.T) {
	loanID := uuid.New()
	dueAt := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		loanID     uuid.UUID
		dueAt      time.Time
		returnedAt time.Time
		wantAmount float64
		wantStatus Status
		wantErr    error
	}{
		{
			name:       "просрочка 1 день",
			loanID:     loanID,
			dueAt:      dueAt,
			returnedAt: dueAt.Add(24 * time.Hour),
			wantAmount: 50.0,
			wantStatus: StatusPending,
			wantErr:    nil,
		},
		{
			name:       "просрочка 1 день + 1 час -> 2 дн",
			loanID:     loanID,
			dueAt:      dueAt,
			returnedAt: dueAt.Add(25 * time.Hour),
			wantAmount: 100.0,
			wantStatus: StatusPending,
			wantErr:    nil,
		},
		{
			name:       "возврат в срок — ошибка",
			loanID:     loanID,
			dueAt:      dueAt,
			returnedAt: dueAt,
			wantErr:    ErrNotOverdue,
		},
		{
			name:       "возврат раньше срока — ошибка",
			loanID:     loanID,
			dueAt:      dueAt,
			returnedAt: dueAt.Add(-time.Hour),
			wantErr:    ErrNotOverdue,
		},
		{
			name:       "пустой loanID",
			loanID:     uuid.Nil,
			dueAt:      dueAt,
			returnedAt: dueAt.Add(time.Hour),
			wantErr:    ErrEmptyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFine(tt.loanID, tt.dueAt, tt.returnedAt)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, f.Status())
			require.Equal(t, tt.wantAmount, f.Amount())
		})
	}
}

package loan

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// loanInStatus возвращает выдачу в заданном статусе.
// Времена подобраны так, что все временные пороги уже пройдены:
// в матрице переходов должна решать только таблица transitions,
// а не проверки TTL/срока.
func loanInStatus(s Status, now time.Time) *Loan {
	past := now.Add(-4 * 24 * time.Hour) // бронь старше TTL (3 дня)
	due := now.Add(-24 * time.Hour)      // срок возврата уже истёк
	return &Loan{
		id:         uuid.New(),
		copyID:     uuid.New(),
		readerID:   uuid.New(),
		status:     s,
		reservedAt: past,
		dueAt:      &due,
	}
}
func TestLoan_Transitions(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC) // фиксированное, не time.Now()

	tests := []struct {
		name    string
		from    Status
		action  func(l *Loan) error
		wantErr error
	}{
		{
			name:    "issue из reserved — успех",
			from:    StatusReserved,
			action:  func(l *Loan) error { return l.Issue(now) },
			wantErr: ErrInvalidTransition, // Сломал ради теста
		},
		{
			name:    "issue из returned — запрещено",
			from:    StatusReturned,
			action:  func(l *Loan) error { return l.Issue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "issue из expired — запрещено",
			from:    StatusExpired,
			action:  func(l *Loan) error { return l.Issue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "issue из issued — запрещено",
			from:    StatusIssued,
			action:  func(l *Loan) error { return l.Issue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "issue из overdue — запрещено",
			from:    StatusOverdue,
			action:  func(l *Loan) error { return l.Issue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "return из reserved — запрещено",
			from:    StatusReserved,
			action:  func(l *Loan) error { return l.Return(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "return из issued — успех",
			from:    StatusIssued,
			action:  func(l *Loan) error { return l.Return(now) },
			wantErr: nil,
		},
		{
			name:    "return из overdue — успех",
			from:    StatusOverdue,
			action:  func(l *Loan) error { return l.Return(now) },
			wantErr: nil,
		},
		{
			name:    "return из expired — запрещено",
			from:    StatusExpired,
			action:  func(l *Loan) error { return l.Return(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "return из returned — запрещено",
			from:    StatusReturned,
			action:  func(l *Loan) error { return l.Return(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "expire из reserved — успех",
			from:    StatusReserved,
			action:  func(l *Loan) error { return l.Expire(now) },
			wantErr: nil,
		},
		{
			name:    "expire из expired — запрещено",
			from:    StatusExpired,
			action:  func(l *Loan) error { return l.Expire(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "expire из issued — запрещено",
			from:    StatusIssued,
			action:  func(l *Loan) error { return l.Expire(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "expire из overdue — запрещено",
			from:    StatusOverdue,
			action:  func(l *Loan) error { return l.Expire(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "expire из returned — запрещено",
			from:    StatusReturned,
			action:  func(l *Loan) error { return l.Expire(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "markOverdue из overdue — запрещено",
			from:    StatusOverdue,
			action:  func(l *Loan) error { return l.MarkOverdue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "markOverdue из issued — успех",
			from:    StatusIssued,
			action:  func(l *Loan) error { return l.MarkOverdue(now) },
			wantErr: nil,
		},
		{
			name:    "markOverdue из reserved — запрещено",
			from:    StatusReserved,
			action:  func(l *Loan) error { return l.MarkOverdue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "markOverdue из expired — запрещено",
			from:    StatusExpired,
			action:  func(l *Loan) error { return l.MarkOverdue(now) },
			wantErr: ErrInvalidTransition,
		},
		{
			name:    "markOverdue из returned — запрещено",
			from:    StatusReturned,
			action:  func(l *Loan) error { return l.MarkOverdue(now) },
			wantErr: ErrInvalidTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := loanInStatus(tt.from, now)

			err := tt.action(l)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.from, l.Status(), "статус не должен меняться при отказе")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestReserve(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC) // фиксированное, не time.Now()
	copyID := uuid.New()
	readerID := uuid.New()

	t.Run("успешное резервирование", func(t *testing.T) {
		l, err := Reserve(copyID, readerID, now)
		require.NoError(t, err)
		require.Equal(t, StatusReserved, l.Status())
		require.Equal(t, now, l.ReservedAt())
		require.NotEqual(t, uuid.Nil, l.ID())
	})

	t.Run("пустой copyID", func(t *testing.T) {
		_, err := Reserve(uuid.Nil, readerID, now)
		require.ErrorIs(t, err, ErrEmptyID)
	})

	t.Run("пустой readerID", func(t *testing.T) {
		_, err := Reserve(copyID, uuid.Nil, now)
		require.ErrorIs(t, err, ErrEmptyID)
	})
}

func TestIssue_SetsDueDate(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	copyID := uuid.New()
	readerID := uuid.New()

	l, err := Reserve(copyID, readerID, now)
	require.NoError(t, err)

	err = l.Issue(now)
	require.NoError(t, err)

	expectedDue := now.AddDate(0, 0, LoanPeriodDays)
	require.Equal(t, expectedDue, *l.DueAt())
}

func TestExpire_TooEarly(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	copyID := uuid.New()
	readerID := uuid.New()

	// Создаём бронь, которой всего 1 день (меньше ReservationTTL = 3 дня)
	l, err := Reserve(copyID, readerID, now.Add(-24*time.Hour))
	require.NoError(t, err)

	err = l.Expire(now)
	require.ErrorIs(t, err, ErrNotYetDue)
	require.Equal(t, StatusReserved, l.Status(), "статус не должен измениться")
}

func TestMarkOverdue_BeforeDue(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	copyID := uuid.New()
	readerID := uuid.New()

	l, err := Reserve(copyID, readerID, now) //Если тестов много, а ошибка будет в Reserve упадут зависящие тесты
	require.NoError(t, err)

	err = l.Issue(now)
	require.NoError(t, err)

	err = l.MarkOverdue(now)
	require.ErrorIs(t, err, ErrNotYetDue)
	require.Equal(t, StatusIssued, l.Status(), "статус не должен измениться")
}

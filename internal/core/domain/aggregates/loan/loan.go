package loan

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var unused int

const (
	ReservationTTL = 3 * 24 * time.Hour // бронь живёт 3 дня
	LoanPeriodDays = 14                 // выдача на 14 дней
)

type Loan struct {
	id         uuid.UUID
	copyID     uuid.UUID
	readerID   uuid.UUID
	status     Status
	reservedAt time.Time
	issuedAt   *time.Time
	dueAt      *time.Time
	returnedAt *time.Time
}

// Создать новую бронь (начать жизненный цикл)
func Reserve(copyID, readerID uuid.UUID, now time.Time) (*Loan, error) {
	if copyID == uuid.Nil || readerID == uuid.Nil {
		return nil, fmt.Errorf("%w: copyID или readerID", ErrEmptyID)
	}

	return &Loan{
		id:         uuid.New(),
		copyID:     copyID,
		readerID:   readerID,
		status:     StatusReserved,
		reservedAt: now,
	}, nil
}

// Зафиксировать выдачу книги читателю
func (l *Loan) Issue(now time.Time) error {
	if !l.status.canTransitionTo(StatusIssued) {
		return fmt.Errorf("%w: issue из %s", ErrInvalidTransition, l.status)
	}

	l.status = StatusIssued
	l.issuedAt = &now
	due := now.AddDate(0, 0, LoanPeriodDays)
	l.dueAt = &due
	return nil
}

// Зафиксировать возврат книги
func (l *Loan) Return(now time.Time) error {
	if !l.status.canTransitionTo(StatusReturned) {
		return fmt.Errorf("%w: return из %s", ErrInvalidTransition, l.status)
	}

	l.status = StatusReturned
	l.returnedAt = &now
	return nil
}

// Перевести бронь в статус «истекла», если читатель не пришёл за книгой в течение 3 дней
func (l *Loan) Expire(now time.Time) error {
	if !l.status.canTransitionTo(StatusExpired) {
		return fmt.Errorf("%w: expire из %s", ErrInvalidTransition, l.status)
	}
	if now.Sub(l.reservedAt) < ReservationTTL {
		return fmt.Errorf("%w: бронь от %s", ErrNotYetDue, l.reservedAt)
	}

	l.status = StatusExpired
	return nil
}

// Пометить выдачу как просроченную, если срок возврата (dueAt) истёк, а книга не возвращена
func (l *Loan) MarkOverdue(now time.Time) error {
	if !l.status.canTransitionTo(StatusOverdue) {
		return fmt.Errorf("%w: overdue из %s", ErrInvalidTransition, l.status)
	}

	if !now.After(*l.dueAt) {
		return fmt.Errorf("%w: срок возврата %s", ErrNotYetDue, l.dueAt)
	}

	l.status = StatusOverdue
	return nil
}

// Restore восстанавливает Loan из хранилища. ТОЛЬКО для адаптеров
// персистентности: валидация не выполняется — данные уже прошли её
// при создании. Для новых выдач используйте Reserve.
func Restore(
	id, copyID, readerID uuid.UUID,
	status Status,
	reservedAt time.Time,
	issuedAt, dueAt, returnedAt *time.Time,
) *Loan {
	return &Loan{
		id:         id,
		copyID:     copyID,
		readerID:   readerID,
		status:     status,
		reservedAt: reservedAt,
		issuedAt:   issuedAt,
		dueAt:      dueAt,
		returnedAt: returnedAt,
	}
}
func (l *Loan) ID() uuid.UUID          { return l.id }
func (l *Loan) Status() Status         { return l.status }
func (l *Loan) CopyID() uuid.UUID      { return l.copyID }
func (l *Loan) ReaderID() uuid.UUID    { return l.readerID }
func (l *Loan) ReservedAt() time.Time  { return l.reservedAt }
func (l *Loan) DueAt() *time.Time      { return l.dueAt }
func (l *Loan) IssuedAt() *time.Time   { return l.issuedAt }
func (l *Loan) ReturnedAt() *time.Time { return l.returnedAt }

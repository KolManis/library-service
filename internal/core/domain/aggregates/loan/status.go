package loan

// Status — статус выдачи в машине переходов Loan.
type Status string

const (
	// StatusReserved — книга забронирована, но ещё не выдана на руки.
	StatusReserved Status = "reserved"
	// StatusIssued — книга выдана читателю на руки.
	StatusIssued Status = "issued"
	// StatusReturned — книга возвращена. Терминальный статус.
	StatusReturned Status = "returned"
	// StatusExpired — бронь не была востребована в течение ReservationTTL. Терминальный статус.
	StatusExpired Status = "expired"
	// StatusOverdue — книга выдана, но срок возврата (due_at) истёк.
	StatusOverdue Status = "overdue"
)

var transitions = map[Status][]Status{
	StatusReserved: {StatusIssued, StatusExpired},
	StatusIssued:   {StatusReturned, StatusOverdue},
	StatusOverdue:  {StatusReturned},
}

// Используется внутри функций для проверки допустимости перехода
func (s Status) canTransitionTo(to Status) bool {
	for _, candidate := range transitions[s] {
		if candidate == to {
			return true
		}
	}

	return false
}

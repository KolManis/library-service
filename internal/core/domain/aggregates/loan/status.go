package loan

type Status string

const (
	StatusReserved Status = "reserved"
	StatusIssued   Status = "issued"
	StatusReturned Status = "returned"
	StatusExpired  Status = "expired"
	StatusOverdue  Status = "overdue"
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

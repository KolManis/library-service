package events

// DomainEvent — маркер: домен поднимает события через этот интерфейс,
// не зная, кто и как их будет публиковать.
type DomainEvent interface {
	EventType() string
}

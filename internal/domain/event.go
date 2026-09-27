package domain

type EventID string

type Event struct {
	ID   EventID
	Name string
}

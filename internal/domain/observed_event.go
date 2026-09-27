package domain

import "time"

type ObservedEvent struct {
	EventID   EventID
	Timestamp time.Time
}

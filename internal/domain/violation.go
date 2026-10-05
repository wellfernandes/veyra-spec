package domain

// Violation describes the first observed event that cannot follow a Specification.
type Violation struct {
	TracePosition  int
	CurrentState   StateID
	ObservedEvent  ObservedEvent
	ExpectedEvents []Event
	Path           []Transition
}

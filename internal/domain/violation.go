package domain

// Violation describes the first observed event that cannot follow a Specification.
type Violation struct {
	TracePosition  int
	CurrentState   StateID
	ObservedEvent  ObservedEvent
	ExpectedEvents []Event
	Path           []Transition
}

// CompletionViolation describes a complete Trace that ends outside the accepted final states.
type CompletionViolation struct {
	CurrentState        StateID
	ExpectedFinalStates []StateID
	Path                []Transition
}

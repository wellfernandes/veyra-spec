package domain

type TransitionID string

type Transition struct {
	ID    TransitionID
	From  StateID
	Event EventID
	To    StateID
}

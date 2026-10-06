package validation

import (
	"errors"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

// ErrUninitializedSpecification is returned when Validate receives a zero-value Specification.
var ErrUninitializedSpecification = errors.New("validation requires a Specification constructed with domain.NewSpecification")

// Result describes the outcome of validating a Trace against a Specification.
type Result struct {
	FinalState domain.StateID
	Violation  *domain.Violation
}

// Valid reports whether the Trace followed the Specification without a violation.
func (r Result) Valid() bool {
	return r.Violation == nil
}

// Validate evaluates the Trace in order and stops at the first missing transition.
func Validate(specification domain.Specification, trace domain.Trace) (Result, error) {
	if specification.ID() == "" {
		return Result{}, ErrUninitializedSpecification
	}

	currentState := specification.InitialState()
	transitionsByState := make(map[domain.StateID][]domain.Transition)
	eventsByID := make(map[domain.EventID]domain.Event)

	for _, event := range specification.Events() {
		eventsByID[event.ID] = event
	}
	for _, transition := range specification.Transitions() {
		transitionsByState[transition.From] = append(
			transitionsByState[transition.From], transition)
	}

	path := make([]domain.Transition, 0, len(trace.ObservedEvents))
	for position, observedEvent := range trace.ObservedEvents {
		transitions := transitionsByState[currentState]
		matchedTransition := -1
		for index, transition := range transitions {
			if transition.Event == observedEvent.EventID {
				matchedTransition = index
				break
			}
		}

		if matchedTransition == -1 {
			expectedEvents := make([]domain.Event, 0, len(transitions))
			for _, transition := range transitions {
				expectedEvents = append(expectedEvents, eventsByID[transition.Event])
			}

			return Result{
				FinalState: currentState,
				Violation: &domain.Violation{
					TracePosition:  position + 1,
					CurrentState:   currentState,
					ObservedEvent:  observedEvent,
					ExpectedEvents: expectedEvents,
					Path:           append([]domain.Transition(nil), path...),
				},
			}, nil
		}

		transition := transitions[matchedTransition]
		path = append(path, transition)
		currentState = transition.To
	}

	return Result{FinalState: currentState}, nil
}

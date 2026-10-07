package validation

import (
	"errors"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

// ErrUninitializedSpecification is returned when Validate receives a zero-value Specification.
var ErrUninitializedSpecification = errors.New("validation requires a Specification constructed with domain.NewSpecification")

// ErrInvalidTraceCompleteness is returned when a Trace has an unknown completeness value.
var ErrInvalidTraceCompleteness = errors.New("trace has an invalid completeness value")

// CompletionStatus describes what validation can conclude about execution completion.
type CompletionStatus uint8

const (
	CompletionStatusUnknown CompletionStatus = iota
	CompletionStatusInProgress
	CompletionStatusNotEvaluated
	CompletionStatusSatisfied
	CompletionStatusFailed
)

// Result describes the outcome of validating a Trace against a Specification.
type Result struct {
	CurrentState        domain.StateID
	CompletionStatus    CompletionStatus
	Violation           *domain.Violation
	CompletionViolation *domain.CompletionViolation
}

// Consistent reports whether all observed events followed valid transitions.
func (r Result) Consistent() bool {
	return r.Violation == nil
}

// Completed reports whether a complete Trace ended in an accepted final state.
func (r Result) Completed() bool {
	return r.CompletionStatus == CompletionStatusSatisfied
}

// Valid reports whether no event or completion violation was found.
// A valid result does not imply that the Trace represents a complete execution.
func (r Result) Valid() bool {
	return r.Consistent() && r.CompletionViolation == nil
}

// Validate evaluates observed events and, for complete Traces, checks the final state when defined.
func Validate(specification domain.Specification, trace domain.Trace) (Result, error) {
	if specification.ID() == "" {
		return Result{}, ErrUninitializedSpecification
	}
	if trace.Completeness > domain.TraceCompletenessComplete {
		return Result{}, ErrInvalidTraceCompleteness
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
				CurrentState:     currentState,
				CompletionStatus: CompletionStatusNotEvaluated,
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

	result := Result{CurrentState: currentState}
	switch trace.Completeness {
	case domain.TraceCompletenessUnknown:
		result.CompletionStatus = CompletionStatusUnknown
	case domain.TraceCompletenessPartial:
		result.CompletionStatus = CompletionStatusInProgress
	case domain.TraceCompletenessComplete:
		finalStates := specification.FinalStates()
		if len(finalStates) == 0 {
			result.CompletionStatus = CompletionStatusNotEvaluated
			break
		}

		for _, finalState := range finalStates {
			if currentState == finalState {
				result.CompletionStatus = CompletionStatusSatisfied
				return result, nil
			}
		}

		result.CompletionStatus = CompletionStatusFailed
		result.CompletionViolation = &domain.CompletionViolation{
			CurrentState:        currentState,
			ExpectedFinalStates: finalStates,
			Path:                append([]domain.Transition(nil), path...),
		}
	}

	return result, nil
}

// Package veyra provides the public Go API for defining and validating
// behavioral Specifications and Traces.
package veyra

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wellfernandes/veyra-spec/internal/domain"
	"github.com/wellfernandes/veyra-spec/internal/validation"
)

// ErrInvalidSpecification identifies a structurally invalid Specification
// definition. The returned error includes all validation details.
var ErrInvalidSpecification = errors.New("invalid specification")

// SpecificationError contains the structural problems found while building a
// Specification. It matches ErrInvalidSpecification and does not expose
// internal domain error types.
type SpecificationError struct {
	problems []string
}

// Error returns the aggregate validation message.
func (err *SpecificationError) Error() string {
	if err == nil || len(err.problems) == 0 {
		return ErrInvalidSpecification.Error()
	}
	return ErrInvalidSpecification.Error() + ": " + strings.Join(err.problems, "; ")
}

// Unwrap identifies this as an invalid Specification error.
func (err *SpecificationError) Unwrap() error {
	return ErrInvalidSpecification
}

// Problems returns the individual structural validation messages.
func (err *SpecificationError) Problems() []string {
	if err == nil {
		return nil
	}
	return append([]string(nil), err.problems...)
}

// ErrUninitializedSpecification identifies a zero-value Specification that
// was not created with NewSpecification.
var ErrUninitializedSpecification = errors.New("uninitialized specification")

// ErrInvalidTraceCompleteness identifies a Trace with an unsupported
// Completeness value.
var ErrInvalidTraceCompleteness = errors.New("invalid trace completeness")

// State describes a state in a Specification. ID is required; Name is
// optional.
type State struct {
	ID   string // Required unique identifier for this State.
	Name string // Optional human-readable name.
}

// Event describes an event in a Specification. ID is required; Name is
// optional.
type Event struct {
	ID   string // Required unique identifier for this Event.
	Name string // Optional human-readable name.
}

// Transition describes a state change in response to an Event.
type Transition struct {
	ID    string // Required unique identifier for this Transition.
	From  string // ID of the source State.
	Event string // ID of the Event that triggers this Transition.
	To    string // ID of the target State.
}

// SpecificationDefinition contains the data used to construct a
// Specification. ID and InitialState are required. State, Event, Transition,
// and FinalStates identifiers must refer to the corresponding definitions.
type SpecificationDefinition struct {
	ID           string       // Required unique identifier for this Specification.
	Name         string       // Optional human-readable name.
	Description  string       // Optional description.
	InitialState string       // Required ID of a State in States.
	States       []State      // States defined by this Specification.
	Events       []Event      // Events defined by this Specification.
	Transitions  []Transition // Valid State and Event combinations.
	FinalStates  []string     // IDs of States that satisfy complete execution.
}

// Specification is a validated behavioral model. Its fields are private;
// construct it with NewSpecification.
type Specification struct {
	value domain.Specification
}

// NewSpecification validates definition and returns an opaque Specification.
// Slices in definition are copied, so subsequent changes to them do not alter
// the returned Specification.
func NewSpecification(definition SpecificationDefinition) (Specification, error) {
	domainDefinition := domain.SpecificationDefinition{
		ID:           domain.SpecificationID(definition.ID),
		Name:         definition.Name,
		Description:  definition.Description,
		InitialState: domain.StateID(definition.InitialState),
		States:       make([]domain.State, len(definition.States)),
		Events:       make([]domain.Event, len(definition.Events)),
		Transitions:  make([]domain.Transition, len(definition.Transitions)),
		FinalStates:  make([]domain.StateID, len(definition.FinalStates)),
	}

	for index, state := range definition.States {
		domainDefinition.States[index] = domain.State{
			ID:   domain.StateID(state.ID),
			Name: state.Name,
		}
	}
	for index, event := range definition.Events {
		domainDefinition.Events[index] = domain.Event{
			ID:   domain.EventID(event.ID),
			Name: event.Name,
		}
	}
	for index, transition := range definition.Transitions {
		domainDefinition.Transitions[index] = domain.Transition{
			ID:    domain.TransitionID(transition.ID),
			From:  domain.StateID(transition.From),
			Event: domain.EventID(transition.Event),
			To:    domain.StateID(transition.To),
		}
	}
	for index, finalState := range definition.FinalStates {
		domainDefinition.FinalStates[index] = domain.StateID(finalState)
	}

	specification, err := domain.NewSpecification(domainDefinition)
	if err != nil {
		var domainErrors *domain.SpecificationErrors
		if errors.As(err, &domainErrors) {
			problems := make([]string, len(domainErrors.Errors))
			for index, problem := range domainErrors.Errors {
				problems[index] = problem.Error()
			}
			return Specification{}, &SpecificationError{problems: problems}
		}
		return Specification{}, &SpecificationError{problems: []string{err.Error()}}
	}

	return Specification{value: specification}, nil
}

// TraceCompleteness indicates how much of an execution a Trace represents.
type TraceCompleteness uint8

const (
	// TraceCompletenessUnknown is the zero value and means that it is unknown
	// whether observation is complete.
	TraceCompletenessUnknown TraceCompleteness = iota
	// TraceCompletenessPartial means that the Trace is a known prefix of an
	// execution.
	TraceCompletenessPartial
	// TraceCompletenessComplete means that the Trace contains the complete
	// execution being validated.
	TraceCompletenessComplete
)

// ObservedEvent records an event in a Trace. Timestamp is optional; its zero
// value means no timestamp was supplied.
type ObservedEvent struct {
	EventID   string    // ID of the observed Event.
	Timestamp time.Time // Optional occurrence time; zero means unspecified.
}

// Trace is an ordered sequence of observed Events. The order of
// ObservedEvents is preserved during validation.
type Trace struct {
	ObservedEvents []ObservedEvent   // Events in execution order.
	Completeness   TraceCompleteness // Whether this Trace represents a full execution.
}

// CompletionStatus describes what validation can conclude about execution
// completion.
type CompletionStatus uint8

const (
	// CompletionStatusUnknown means the Trace's completeness is unknown.
	CompletionStatusUnknown CompletionStatus = iota
	// CompletionStatusInProgress means the Trace is partial and consistent.
	CompletionStatusInProgress
	// CompletionStatusNotEvaluated means completion could not be evaluated,
	// either because the Trace has no final-state requirement or an event
	// violation occurred.
	CompletionStatusNotEvaluated
	// CompletionStatusSatisfied means a complete Trace ended in a final State.
	CompletionStatusSatisfied
	// CompletionStatusFailed means a complete Trace did not end in a final State.
	CompletionStatusFailed
)

// Violation describes the first observed event that cannot follow a
// Specification. TracePosition is one-based. Path contains the transitions
// that successfully preceded the violation.
type Violation struct {
	TracePosition  int           // One-based position of the first invalid event.
	CurrentState   string        // State before the invalid event.
	ObservedEvent  ObservedEvent // Event that could not follow the Specification.
	ExpectedEvents []Event       // Events with valid transitions from CurrentState.
	Path           []Transition  // Valid transitions before the violation.
}

// CompletionViolation describes a complete Trace that ended outside the
// accepted final states.
type CompletionViolation struct {
	CurrentState        string       // State where the complete Trace ended.
	ExpectedFinalStates []string     // States accepted as complete by the Specification.
	Path                []Transition // Transitions followed by the Trace.
}

// Result describes the outcome of validating a Trace against a Specification.
// Its diagnostic slices are independent copies and may be modified by callers.
type Result struct {
	CurrentState        string               // State reached after the processed Trace.
	CompletionStatus    CompletionStatus     // Conclusion about execution completion.
	Violation           *Violation           // First event divergence, if any.
	CompletionViolation *CompletionViolation // Final-state divergence, if any.
}

// Consistent reports whether all observed events followed valid transitions.
func (result Result) Consistent() bool {
	return result.Violation == nil
}

// Completed reports whether a complete Trace ended in an accepted final State.
func (result Result) Completed() bool {
	return result.CompletionStatus == CompletionStatusSatisfied
}

// Valid reports whether no event or completion violation was found. A valid
// result does not imply that the Trace represents a complete execution.
func (result Result) Valid() bool {
	return result.Consistent() && result.CompletionViolation == nil
}

// Validate evaluates observed events and, for complete Traces, checks the
// final State when the Specification defines final states. Behavioral
// violations are returned in Result; error is reserved for invalid input or
// an execution failure.
func Validate(specification Specification, trace Trace) (Result, error) {
	domainTrace := domain.Trace{
		Completeness:   domain.TraceCompleteness(trace.Completeness),
		ObservedEvents: make([]domain.ObservedEvent, len(trace.ObservedEvents)),
	}
	for index, observedEvent := range trace.ObservedEvents {
		domainTrace.ObservedEvents[index] = domain.ObservedEvent{
			EventID:   domain.EventID(observedEvent.EventID),
			Timestamp: observedEvent.Timestamp,
		}
	}

	domainResult, err := validation.Validate(specification.value, domainTrace)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrUninitializedSpecification):
			return Result{}, fmt.Errorf("%w: %v", ErrUninitializedSpecification, err)
		case errors.Is(err, validation.ErrInvalidTraceCompleteness):
			return Result{}, fmt.Errorf("%w: %v", ErrInvalidTraceCompleteness, err)
		default:
			return Result{}, fmt.Errorf("validate specification: %v", err)
		}
	}

	completionStatus, err := publicCompletionStatus(domainResult.CompletionStatus)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		CurrentState:     string(domainResult.CurrentState),
		CompletionStatus: completionStatus,
	}

	if domainViolation := domainResult.Violation; domainViolation != nil {
		violation := &Violation{
			TracePosition: domainViolation.TracePosition,
			CurrentState:  string(domainViolation.CurrentState),
			ObservedEvent: ObservedEvent{
				EventID:   string(domainViolation.ObservedEvent.EventID),
				Timestamp: domainViolation.ObservedEvent.Timestamp,
			},
			ExpectedEvents: make([]Event, len(domainViolation.ExpectedEvents)),
			Path:           make([]Transition, len(domainViolation.Path)),
		}
		for index, event := range domainViolation.ExpectedEvents {
			violation.ExpectedEvents[index] = Event{ID: string(event.ID), Name: event.Name}
		}
		for index, transition := range domainViolation.Path {
			violation.Path[index] = publicTransition(transition)
		}
		result.Violation = violation
	}

	if domainViolation := domainResult.CompletionViolation; domainViolation != nil {
		violation := &CompletionViolation{
			CurrentState:        string(domainViolation.CurrentState),
			ExpectedFinalStates: make([]string, len(domainViolation.ExpectedFinalStates)),
			Path:                make([]Transition, len(domainViolation.Path)),
		}
		for index, finalState := range domainViolation.ExpectedFinalStates {
			violation.ExpectedFinalStates[index] = string(finalState)
		}
		for index, transition := range domainViolation.Path {
			violation.Path[index] = publicTransition(transition)
		}
		result.CompletionViolation = violation
	}

	return result, nil
}

func publicTransition(transition domain.Transition) Transition {
	return Transition{
		ID:    string(transition.ID),
		From:  string(transition.From),
		Event: string(transition.Event),
		To:    string(transition.To),
	}
}

func publicCompletionStatus(status validation.CompletionStatus) (CompletionStatus, error) {
	switch status {
	case validation.CompletionStatusUnknown:
		return CompletionStatusUnknown, nil
	case validation.CompletionStatusInProgress:
		return CompletionStatusInProgress, nil
	case validation.CompletionStatusNotEvaluated:
		return CompletionStatusNotEvaluated, nil
	case validation.CompletionStatusSatisfied:
		return CompletionStatusSatisfied, nil
	case validation.CompletionStatusFailed:
		return CompletionStatusFailed, nil
	default:
		return CompletionStatusUnknown, fmt.Errorf("validation returned an unsupported completion status: %d", status)
	}
}

package domain

import (
	"fmt"
	"strings"
)

// SpecificationID uniquely identifies a behavioral specification.
type SpecificationID string

// SpecificationDefinition contains the data used to create a Specification.
// It is an input structure, not a validated domain object.
type SpecificationDefinition struct {
	ID           SpecificationID
	Name         string
	Description  string
	InitialState StateID
	States       []State
	Events       []Event
	Transitions  []Transition
}

// SpecificationErrors contains all structural errors found while
// constructing a Specification.
type SpecificationErrors struct {
	Errors []error
}

func (e *SpecificationErrors) Error() string {
	messages := make([]string, 0, len(e.Errors))

	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}

	return "invalid specification: " + strings.Join(messages, "; ")
}

// Specification represents a validated behavioral model.
//
// Its fields are private to prevent callers from modifying the model
// without going through its construction rules.
type Specification struct {
	id           SpecificationID
	name         string
	description  string
	initialState StateID
	states       []State
	events       []Event
	transitions  []Transition
}

// NewSpecification validates the definition and returns a Specification
// only if all structural invariants are satisfied.
func NewSpecification(def SpecificationDefinition) (Specification, error) {
	var validationErrors []error

	if strings.TrimSpace(string(def.ID)) == "" {
		validationErrors = append(validationErrors,
			fmt.Errorf("specification ID must not be empty"))
	}

	statesByID := make(map[StateID]struct{}, len(def.States))
	for i, state := range def.States {
		if strings.TrimSpace(string(state.ID)) == "" {
			validationErrors = append(validationErrors,
				fmt.Errorf("state at index %d has an empty ID", i))
			continue
		}

		if _, exists := statesByID[state.ID]; exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("duplicate state ID %q", state.ID))
			continue
		}

		statesByID[state.ID] = struct{}{}
	}

	if strings.TrimSpace(string(def.InitialState)) == "" {
		validationErrors = append(validationErrors,
			fmt.Errorf("initial state ID must not be empty"))
	} else if _, exists := statesByID[def.InitialState]; !exists {
		validationErrors = append(validationErrors,
			fmt.Errorf("initial state %q does not reference an existing state", def.InitialState))
	}

	eventsByID := make(map[EventID]struct{}, len(def.Events))
	for i, event := range def.Events {
		if strings.TrimSpace(string(event.ID)) == "" {
			validationErrors = append(validationErrors,
				fmt.Errorf("event at index %d has an empty ID", i))
			continue
		}

		if _, exists := eventsByID[event.ID]; exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("duplicate event ID %q", event.ID))
			continue
		}

		eventsByID[event.ID] = struct{}{}
	}

	transitionIDs := make(map[TransitionID]struct{}, len(def.Transitions))
	type transitionKey struct {
		from  StateID
		event EventID
	}
	transitionPairs := make(map[transitionKey]struct{}, len(def.Transitions))

	for i, transition := range def.Transitions {
		if strings.TrimSpace(string(transition.ID)) == "" {
			validationErrors = append(validationErrors,
				fmt.Errorf("transition at index %d has an empty ID", i))
		} else {
			if _, exists := transitionIDs[transition.ID]; exists {
				validationErrors = append(validationErrors,
					fmt.Errorf("duplicate transition ID %q", transition.ID))
			} else {
				transitionIDs[transition.ID] = struct{}{}
			}
		}

		if _, exists := statesByID[transition.From]; !exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("transition %q references unknown source state %q",
					transition.ID, transition.From))
		}

		if _, exists := statesByID[transition.To]; !exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("transition %q references unknown target state %q",
					transition.ID, transition.To))
		}

		if _, exists := eventsByID[transition.Event]; !exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("transition %q references unknown event %q",
					transition.ID, transition.Event))
		}

		pair := transitionKey{from: transition.From, event: transition.Event}
		if _, exists := transitionPairs[pair]; exists {
			validationErrors = append(validationErrors,
				fmt.Errorf("multiple transitions defined for state %q and event %q",
					transition.From, transition.Event))
		} else {
			transitionPairs[pair] = struct{}{}
		}
	}

	if len(validationErrors) > 0 {
		return Specification{}, &SpecificationErrors{Errors: validationErrors}
	}

	// Copy slices so later changes to the definition's slices do not
	// mutate the constructed Specification.
	return Specification{
		id:           def.ID,
		name:         def.Name,
		description:  def.Description,
		initialState: def.InitialState,
		states:       append([]State(nil), def.States...),
		events:       append([]Event(nil), def.Events...),
		transitions:  append([]Transition(nil), def.Transitions...),
	}, nil
}

// ID returns the specification's identifier.
func (s Specification) ID() SpecificationID {
	return s.id
}

// Name returns the specification's human-readable name.
func (s Specification) Name() string {
	return s.name
}

// Description returns the specification's description.
func (s Specification) Description() string {
	return s.description
}

// InitialState returns the ID of the specification's initial state.
func (s Specification) InitialState() StateID {
	return s.initialState
}

// States returns a copy of the specification's states.
func (s Specification) States() []State {
	return append([]State(nil), s.states...)
}

// Events returns a copy of the specification's events.
func (s Specification) Events() []Event {
	return append([]Event(nil), s.events...)
}

// Transitions returns a copy of the specification's transitions.
func (s Specification) Transitions() []Transition {
	return append([]Transition(nil), s.transitions...)
}

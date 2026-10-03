package domain

import (
	"errors"
	"strings"
	"testing"
)

func validSpecificationDefinition() SpecificationDefinition {
	return SpecificationDefinition{
		ID:           "payment-flow",
		Name:         "Payment Flow",
		Description:  "Behavioral model for a payment process.",
		InitialState: "pending",
		States: []State{
			{ID: "pending", Name: "Pending"},
			{ID: "approved", Name: "Approved"},
			{ID: "rejected", Name: "Rejected"},
		},
		Events: []Event{
			{ID: "payment_approved", Name: "Payment Approved"},
			{ID: "payment_rejected", Name: "Payment Rejected"},
		},
		Transitions: []Transition{
			{
				ID:    "approve-payment",
				From:  "pending",
				Event: "payment_approved",
				To:    "approved",
			},
			{
				ID:    "reject-payment",
				From:  "pending",
				Event: "payment_rejected",
				To:    "rejected",
			},
		},
	}
}

func TestNewSpecification_ValidDefinition(t *testing.T) {
	def := validSpecificationDefinition()

	spec, err := NewSpecification(def)
	if err != nil {
		t.Fatalf("NewSpecification() returned unexpected error: %v", err)
	}

	if got, want := spec.ID(), def.ID; got != want {
		t.Errorf("ID() = %q, want %q", got, want)
	}

	if got, want := spec.Name(), def.Name; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}

	if got, want := spec.Description(), def.Description; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}

	if got, want := spec.InitialState(), def.InitialState; got != want {
		t.Errorf("InitialState() = %q, want %q", got, want)
	}

	if got, want := len(spec.States()), len(def.States); got != want {
		t.Errorf("States() length = %d, want %d", got, want)
	}

	if got, want := len(spec.Events()), len(def.Events); got != want {
		t.Errorf("Events() length = %d, want %d", got, want)
	}

	if got, want := len(spec.Transitions()), len(def.Transitions); got != want {
		t.Errorf("Transitions() length = %d, want %d", got, want)
	}
}

func TestNewSpecification_RejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*SpecificationDefinition)
		wantInError string
	}{
		{
			name: "empty specification ID",
			mutate: func(def *SpecificationDefinition) {
				def.ID = ""
			},
			wantInError: "specification ID must not be empty",
		},
		{
			name: "whitespace specification ID",
			mutate: func(def *SpecificationDefinition) {
				def.ID = "   "
			},
			wantInError: "specification ID must not be empty",
		},
		{
			name: "empty state ID",
			mutate: func(def *SpecificationDefinition) {
				def.States[0].ID = ""
			},
			wantInError: "state at index 0 has an empty ID",
		},
		{
			name: "duplicate state ID",
			mutate: func(def *SpecificationDefinition) {
				def.States = append(def.States, State{
					ID:   "pending",
					Name: "Another Pending",
				})
			},
			wantInError: `duplicate state ID "pending"`,
		},
		{
			name: "empty initial state ID",
			mutate: func(def *SpecificationDefinition) {
				def.InitialState = ""
			},
			wantInError: "initial state ID must not be empty",
		},
		{
			name: "initial state does not exist",
			mutate: func(def *SpecificationDefinition) {
				def.InitialState = "unknown"
			},
			wantInError: `initial state "unknown" does not reference an existing state`,
		},
		{
			name: "empty event ID",
			mutate: func(def *SpecificationDefinition) {
				def.Events[0].ID = ""
			},
			wantInError: "event at index 0 has an empty ID",
		},
		{
			name: "duplicate event ID",
			mutate: func(def *SpecificationDefinition) {
				def.Events = append(def.Events, Event{
					ID:   "payment_approved",
					Name: "Another Approved Event",
				})
			},
			wantInError: `duplicate event ID "payment_approved"`,
		},
		{
			name: "empty transition ID",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions[0].ID = ""
			},
			wantInError: "transition at index 0 has an empty ID",
		},
		{
			name: "duplicate transition ID",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions[1].ID = def.Transitions[0].ID
			},
			wantInError: `duplicate transition ID "approve-payment"`,
		},
		{
			name: "unknown source state",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions[0].From = "unknown"
			},
			wantInError: `references unknown source state "unknown"`,
		},
		{
			name: "unknown target state",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions[0].To = "unknown"
			},
			wantInError: `references unknown target state "unknown"`,
		},
		{
			name: "unknown event",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions[0].Event = "unknown"
			},
			wantInError: `references unknown event "unknown"`,
		},
		{
			name: "non-deterministic transitions",
			mutate: func(def *SpecificationDefinition) {
				def.Transitions = append(def.Transitions, Transition{
					ID:    "approve-payment-again",
					From:  "pending",
					Event: "payment_approved",
					To:    "rejected",
				})
			},
			wantInError: "multiple transitions defined for state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := validSpecificationDefinition()
			tt.mutate(&def)

			_, err := NewSpecification(def)
			if err == nil {
				t.Fatal("NewSpecification() returned nil error, want validation error")
			}

			var specificationErrors *SpecificationErrors
			if !errors.As(err, &specificationErrors) {
				t.Fatalf("error type = %T, want *SpecificationErrors", err)
			}

			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Errorf("error = %q, want it to contain %q",
					err.Error(), tt.wantInError)
			}
		})
	}
}

func TestNewSpecification_AllowsRepeatedNamesAndUnusedDefinitions(t *testing.T) {
	def := validSpecificationDefinition()

	// Repeated names are allowed because IDs define identity.
	def.States[1].Name = def.States[0].Name
	def.Events[1].Name = def.Events[0].Name

	// An unused state and event do not make the model structurally invalid.
	def.States = append(def.States, State{
		ID:   "archived",
		Name: "Archived",
	})
	def.Events = append(def.Events, Event{
		ID:   "payment_refunded",
		Name: "Payment Refunded",
	})

	_, err := NewSpecification(def)
	if err != nil {
		t.Fatalf("NewSpecification() returned unexpected error: %v", err)
	}
}

func TestNewSpecification_ReturnedSlicesDoNotMutateSpecification(t *testing.T) {
	spec, err := NewSpecification(validSpecificationDefinition())
	if err != nil {
		t.Fatalf("NewSpecification() returned unexpected error: %v", err)
	}

	states := spec.States()
	events := spec.Events()
	transitions := spec.Transitions()

	states[0].Name = "Changed State"
	events[0].Name = "Changed Event"
	transitions[0].To = "changed-target"

	if got, want := spec.States()[0].Name, "Pending"; got != want {
		t.Errorf("state name changed through returned slice: got %q, want %q", got, want)
	}

	if got, want := spec.Events()[0].Name, "Payment Approved"; got != want {
		t.Errorf("event name changed through returned slice: got %q, want %q", got, want)
	}

	if got, want := spec.Transitions()[0].To, StateID("approved"); got != want {
		t.Errorf("transition target changed through returned slice: got %q, want %q", got, want)
	}
}

func TestNewSpecification_AggregatesValidationErrors(t *testing.T) {
	def := validSpecificationDefinition()
	def.ID = ""
	def.InitialState = "missing"
	def.Transitions[0].Event = "unknown-event"

	_, err := NewSpecification(def)
	if err == nil {
		t.Fatal("NewSpecification() returned nil error, want validation errors")
	}

	var specificationErrors *SpecificationErrors
	if !errors.As(err, &specificationErrors) {
		t.Fatalf("error type = %T, want *SpecificationErrors", err)
	}

	if got, want := len(specificationErrors.Errors), 3; got != want {
		t.Errorf("number of validation errors = %d, want %d",
			got, want)
	}
}

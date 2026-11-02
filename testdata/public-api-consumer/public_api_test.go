package consumer_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	veyra "github.com/wellfernandes/veyra-spec"
)

func paymentDefinition() veyra.SpecificationDefinition {
	return veyra.SpecificationDefinition{
		ID:           "payment-flow",
		Name:         "Payment Flow",
		InitialState: "pending",
		States: []veyra.State{
			{ID: "pending", Name: "Pending"},
			{ID: "processing", Name: "Processing"},
			{ID: "approved", Name: "Approved"},
			{ID: "rejected", Name: "Rejected"},
		},
		Events: []veyra.Event{
			{ID: "payment_requested", Name: "Payment Requested"},
			{ID: "payment_approved", Name: "Payment Approved"},
			{ID: "payment_rejected", Name: "Payment Rejected"},
			{ID: "payment_cancelled", Name: "Payment Cancelled"},
		},
		Transitions: []veyra.Transition{
			{ID: "start-payment", From: "pending", Event: "payment_requested", To: "processing"},
			{ID: "approve-payment", From: "processing", Event: "payment_approved", To: "approved"},
			{ID: "reject-payment", From: "processing", Event: "payment_rejected", To: "rejected"},
		},
		FinalStates: []string{"approved", "rejected"},
	}
}

func newPaymentSpecification(t *testing.T) veyra.Specification {
	t.Helper()

	specification, err := veyra.NewSpecification(paymentDefinition())
	if err != nil {
		t.Fatalf("NewSpecification() error = %v", err)
	}
	return specification
}

func check() (bool, error) {
	specification, err := veyra.NewSpecification(veyra.SpecificationDefinition{
		ID:           "payment-flow",
		InitialState: "pending",
		States: []veyra.State{
			{ID: "pending"},
			{ID: "approved"},
		},
		Events: []veyra.Event{
			{ID: "payment_approved"},
		},
		Transitions: []veyra.Transition{
			{ID: "approve-payment", From: "pending", Event: "payment_approved", To: "approved"},
		},
		FinalStates: []string{"approved"},
	})
	if err != nil {
		return false, err
	}

	result, err := veyra.Validate(specification, veyra.Trace{
		Completeness: veyra.TraceCompletenessComplete,
		ObservedEvents: []veyra.ObservedEvent{
			{EventID: "payment_approved"},
		},
	})
	if err != nil {
		return false, err
	}
	return result.Valid(), nil
}

func TestREADMEExample(t *testing.T) {
	valid, err := check()
	if err != nil {
		t.Fatalf("check() error = %v", err)
	}
	if !valid {
		t.Fatal("check() = false, want true")
	}
}

func TestPublicAPIValidCompleteTrace(t *testing.T) {
	result, err := veyra.Validate(newPaymentSpecification(t), veyra.Trace{
		Completeness: veyra.TraceCompletenessComplete,
		ObservedEvents: []veyra.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_approved"},
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid() || !result.Consistent() || !result.Completed() {
		t.Fatalf("result = %#v, want valid completed trace", result)
	}
	if got, want := result.CurrentState, "approved"; got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}
	if got, want := result.CompletionStatus, veyra.CompletionStatusSatisfied; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
}

func TestPublicAPIEventViolationPreservesDiagnostics(t *testing.T) {
	timestamp := time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)
	result, err := veyra.Validate(newPaymentSpecification(t), veyra.Trace{
		ObservedEvents: []veyra.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_cancelled", Timestamp: timestamp},
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Valid() || result.Violation == nil {
		t.Fatalf("result = %#v, want event violation", result)
	}

	violation := result.Violation
	if got, want := violation.TracePosition, 2; got != want {
		t.Errorf("TracePosition = %d, want %d", got, want)
	}
	if got, want := violation.CurrentState, "processing"; got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}
	if got, want := violation.ObservedEvent.EventID, "payment_cancelled"; got != want {
		t.Errorf("ObservedEvent.EventID = %q, want %q", got, want)
	}
	if !violation.ObservedEvent.Timestamp.Equal(timestamp) {
		t.Errorf("ObservedEvent.Timestamp = %v, want %v", violation.ObservedEvent.Timestamp, timestamp)
	}
	if got, want := len(violation.ExpectedEvents), 2; got != want {
		t.Fatalf("len(ExpectedEvents) = %d, want %d", got, want)
	}
	if got, want := violation.ExpectedEvents[0].ID, "payment_approved"; got != want {
		t.Errorf("ExpectedEvents[0].ID = %q, want %q", got, want)
	}
	if got, want := len(violation.Path), 1; got != want {
		t.Fatalf("len(Path) = %d, want %d", got, want)
	}
	if got, want := violation.Path[0].ID, "start-payment"; got != want {
		t.Errorf("Path[0].ID = %q, want %q", got, want)
	}
}

func TestPublicAPICompleteTraceOutsideFinalState(t *testing.T) {
	result, err := veyra.Validate(newPaymentSpecification(t), veyra.Trace{
		Completeness:   veyra.TraceCompletenessComplete,
		ObservedEvents: []veyra.ObservedEvent{{EventID: "payment_requested"}},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Consistent() || result.Valid() || result.Completed() {
		t.Fatalf("result = %#v, want consistent but incomplete result", result)
	}
	if got, want := result.CompletionStatus, veyra.CompletionStatusFailed; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
	if result.CompletionViolation == nil {
		t.Fatal("CompletionViolation = nil, want completion diagnostics")
	}
	if got, want := result.CompletionViolation.CurrentState, "processing"; got != want {
		t.Errorf("CompletionViolation.CurrentState = %q, want %q", got, want)
	}
	if got, want := result.CompletionViolation.ExpectedFinalStates[0], "approved"; got != want {
		t.Errorf("ExpectedFinalStates[0] = %q, want %q", got, want)
	}
	if got, want := result.CompletionViolation.Path[0].ID, "start-payment"; got != want {
		t.Errorf("Path[0].ID = %q, want %q", got, want)
	}
}

func TestPublicAPIMapsPartialAndUnknownCompleteness(t *testing.T) {
	tests := []struct {
		name       string
		trace      veyra.Trace
		wantStatus veyra.CompletionStatus
	}{
		{
			name: "partial",
			trace: veyra.Trace{
				Completeness:   veyra.TraceCompletenessPartial,
				ObservedEvents: []veyra.ObservedEvent{{EventID: "payment_requested"}},
			},
			wantStatus: veyra.CompletionStatusInProgress,
		},
		{
			name: "unknown zero value",
			trace: veyra.Trace{
				ObservedEvents: []veyra.ObservedEvent{{EventID: "payment_requested"}},
			},
			wantStatus: veyra.CompletionStatusUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := veyra.Validate(newPaymentSpecification(t), test.trace)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !result.Valid() || result.Completed() {
				t.Fatalf("result = %#v, want valid non-complete result", result)
			}
			if got := result.CompletionStatus; got != test.wantStatus {
				t.Errorf("CompletionStatus = %v, want %v", got, test.wantStatus)
			}
		})
	}
}

func TestPublicAPIErrorsAreInspectable(t *testing.T) {
	_, err := veyra.NewSpecification(veyra.SpecificationDefinition{})
	if !errors.Is(err, veyra.ErrInvalidSpecification) {
		t.Errorf("NewSpecification() error = %v, want ErrInvalidSpecification", err)
	}
	var specificationError *veyra.SpecificationError
	if !errors.As(err, &specificationError) {
		t.Fatalf("NewSpecification() error type = %T, want *veyra.SpecificationError", err)
	}
	problems := specificationError.Problems()
	for _, detail := range []string{"specification ID must not be empty", "initial state ID must not be empty"} {
		found := false
		for _, problem := range problems {
			if strings.Contains(problem, detail) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SpecificationError.Problems() = %q, want detail %q", problems, detail)
		}
	}
	problems[0] = "changed"
	if got := specificationError.Problems()[0]; got == "changed" {
		t.Error("Problems() returned a slice that aliases the error's stored problems")
	}

	specification := newPaymentSpecification(t)
	_, err = veyra.Validate(veyra.Specification{}, veyra.Trace{})
	if !errors.Is(err, veyra.ErrUninitializedSpecification) {
		t.Errorf("Validate() error = %v, want ErrUninitializedSpecification", err)
	}

	_, err = veyra.Validate(specification, veyra.Trace{
		Completeness: veyra.TraceCompleteness(255),
	})
	if !errors.Is(err, veyra.ErrInvalidTraceCompleteness) {
		t.Errorf("Validate() error = %v, want ErrInvalidTraceCompleteness", err)
	}
}

func TestPublicAPISlicesAreIndependent(t *testing.T) {
	definition := paymentDefinition()
	specification, err := veyra.NewSpecification(definition)
	if err != nil {
		t.Fatalf("NewSpecification() error = %v", err)
	}

	definition.States[0].ID = "changed"
	definition.Events[0].ID = "changed"
	definition.Transitions[0].To = "changed"
	definition.FinalStates[0] = "changed"

	result, err := veyra.Validate(specification, veyra.Trace{
		Completeness:   veyra.TraceCompletenessComplete,
		ObservedEvents: []veyra.ObservedEvent{{EventID: "payment_requested"}},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.CompletionViolation == nil {
		t.Fatal("CompletionViolation = nil, want completion diagnostics")
	}

	result.CompletionViolation.ExpectedFinalStates[0] = "changed"
	result.CompletionViolation.Path[0].ID = "changed"

	again, err := veyra.Validate(specification, veyra.Trace{
		Completeness:   veyra.TraceCompletenessComplete,
		ObservedEvents: []veyra.ObservedEvent{{EventID: "payment_requested"}},
	})
	if err != nil {
		t.Fatalf("second Validate() error = %v", err)
	}
	if got, want := again.CompletionViolation.ExpectedFinalStates[0], "approved"; got != want {
		t.Errorf("ExpectedFinalStates[0] after result mutation = %q, want %q", got, want)
	}
	if got, want := again.CompletionViolation.Path[0].ID, "start-payment"; got != want {
		t.Errorf("Path[0].ID after result mutation = %q, want %q", got, want)
	}

	observedEvents := []veyra.ObservedEvent{
		{EventID: "payment_requested"},
		{EventID: "payment_cancelled"},
	}
	eventResult, err := veyra.Validate(specification, veyra.Trace{ObservedEvents: observedEvents})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	observedEvents[1].EventID = "changed"
	if got, want := eventResult.Violation.ObservedEvent.EventID, "payment_cancelled"; got != want {
		t.Errorf("ObservedEvent.EventID after input mutation = %q, want %q", got, want)
	}

	eventResult.Violation.ExpectedEvents[0].ID = "changed"
	eventResult.Violation.Path[0].ID = "changed"
	eventAgain, err := veyra.Validate(specification, veyra.Trace{
		ObservedEvents: []veyra.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_cancelled"},
		},
	})
	if err != nil {
		t.Fatalf("second event Validate() error = %v", err)
	}
	if got, want := eventAgain.Violation.ExpectedEvents[0].ID, "payment_approved"; got != want {
		t.Errorf("ExpectedEvents[0] after result mutation = %q, want %q", got, want)
	}
	if got, want := eventAgain.Violation.Path[0].ID, "start-payment"; got != want {
		t.Errorf("Path[0].ID after result mutation = %q, want %q", got, want)
	}
}

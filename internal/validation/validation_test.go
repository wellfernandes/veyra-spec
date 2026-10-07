package validation

import (
	"errors"
	"testing"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

func paymentSpecification(t *testing.T) domain.Specification {
	t.Helper()

	specification, err := domain.NewSpecification(domain.SpecificationDefinition{
		ID:           "payment-flow",
		Name:         "Payment Flow",
		InitialState: "pending",
		FinalStates:  []domain.StateID{"approved", "rejected"},
		States: []domain.State{
			{ID: "pending", Name: "Pending"},
			{ID: "processing", Name: "Processing"},
			{ID: "approved", Name: "Approved"},
			{ID: "rejected", Name: "Rejected"},
		},
		Events: []domain.Event{
			{ID: "payment_requested", Name: "Payment Requested"},
			{ID: "payment_approved", Name: "Payment Approved"},
			{ID: "payment_rejected", Name: "Payment Rejected"},
		},
		Transitions: []domain.Transition{
			{ID: "start-payment", From: "pending", Event: "payment_requested", To: "processing"},
			{ID: "approve-payment", From: "processing", Event: "payment_approved", To: "approved"},
			{ID: "reject-payment", From: "processing", Event: "payment_rejected", To: "rejected"},
		},
	})
	if err != nil {
		t.Fatalf("NewSpecification() returned unexpected error: %v", err)
	}

	return specification
}

func TestValidate_ValidTrace(t *testing.T) {
	result, err := Validate(paymentSpecification(t), domain.Trace{
		Completeness: domain.TraceCompletenessComplete,
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_approved"},
		},
	})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !result.Valid() {
		t.Fatalf("Valid() = false, violation = %#v", result.Violation)
	}
	if got, want := result.CurrentState, domain.StateID("approved"); got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}
	if !result.Completed() {
		t.Errorf("Completed() = false, want true; status = %v", result.CompletionStatus)
	}
}

func TestValidate_PartialTraceEndsAtCurrentState(t *testing.T) {
	result, err := Validate(paymentSpecification(t), domain.Trace{
		Completeness: domain.TraceCompletenessPartial,
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
		},
	})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !result.Consistent() || !result.Valid() {
		t.Fatalf("partial Trace should be consistent, result = %#v", result)
	}
	if result.Completed() {
		t.Fatal("Completed() = true for a partial Trace, want false")
	}
	if got, want := result.CompletionStatus, CompletionStatusInProgress; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
	if got, want := result.CurrentState, domain.StateID("processing"); got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}
}

func TestValidate_UnknownTraceCompletenessDoesNotImplyCompletion(t *testing.T) {
	result, err := Validate(paymentSpecification(t), domain.Trace{})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !result.Consistent() || !result.Valid() {
		t.Fatalf("unknown Trace should remain consistent, result = %#v", result)
	}
	if result.Completed() {
		t.Fatal("Completed() = true when Trace completeness is unknown")
	}
	if got, want := result.CompletionStatus, CompletionStatusUnknown; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
}

func TestValidate_CompleteTraceOutsideFinalStateReportsCompletionViolation(t *testing.T) {
	result, err := Validate(paymentSpecification(t), domain.Trace{
		Completeness: domain.TraceCompletenessComplete,
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
		},
	})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if !result.Consistent() {
		t.Fatalf("Consistent() = false, want true; event violation = %#v", result.Violation)
	}
	if result.Valid() || result.Completed() {
		t.Fatalf("complete Trace ending outside final state reported success: %#v", result)
	}
	if got, want := result.CompletionStatus, CompletionStatusFailed; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
	if result.CompletionViolation == nil {
		t.Fatal("CompletionViolation = nil, want completion failure details")
	}
	if got, want := result.CompletionViolation.CurrentState, domain.StateID("processing"); got != want {
		t.Errorf("CompletionViolation.CurrentState = %q, want %q", got, want)
	}
	if got, want := len(result.CompletionViolation.ExpectedFinalStates), 2; got != want {
		t.Fatalf("len(CompletionViolation.ExpectedFinalStates) = %d, want %d", got, want)
	}
	if got, want := result.CompletionViolation.ExpectedFinalStates[0], domain.StateID("approved"); got != want {
		t.Errorf("ExpectedFinalStates[0] = %q, want %q", got, want)
	}
	if got, want := len(result.CompletionViolation.Path), 1; got != want {
		t.Fatalf("len(CompletionViolation.Path) = %d, want %d", got, want)
	}
}

func TestValidate_CompleteTraceWithoutFinalStatesDoesNotAssessCompletion(t *testing.T) {
	specification, err := domain.NewSpecification(domain.SpecificationDefinition{
		ID:           "ongoing-process",
		InitialState: "running",
		States:       []domain.State{{ID: "running", Name: "Running"}},
	})
	if err != nil {
		t.Fatalf("NewSpecification() returned unexpected error: %v", err)
	}

	result, err := Validate(specification, domain.Trace{Completeness: domain.TraceCompletenessComplete})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
	if !result.Valid() || result.Completed() {
		t.Fatalf("completion should be unassessed without final states: %#v", result)
	}
	if got, want := result.CompletionStatus, CompletionStatusNotEvaluated; got != want {
		t.Errorf("CompletionStatus = %v, want %v", got, want)
	}
}

func TestValidate_ReportsFirstDivergenceWithContext(t *testing.T) {
	result, err := Validate(paymentSpecification(t), domain.Trace{
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_cancelled"},
			{EventID: "payment_approved"},
		},
	})
	if err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}

	if result.Valid() {
		t.Fatal("Valid() = true, want a violation")
	}
	if got, want := result.CurrentState, domain.StateID("processing"); got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}

	violation := result.Violation
	if got, want := violation.TracePosition, 2; got != want {
		t.Errorf("TracePosition = %d, want %d", got, want)
	}
	if got, want := violation.CurrentState, domain.StateID("processing"); got != want {
		t.Errorf("CurrentState = %q, want %q", got, want)
	}
	if got, want := violation.ObservedEvent.EventID, domain.EventID("payment_cancelled"); got != want {
		t.Errorf("ObservedEvent.EventID = %q, want %q", got, want)
	}
	if got, want := len(violation.ExpectedEvents), 2; got != want {
		t.Fatalf("len(ExpectedEvents) = %d, want %d", got, want)
	}
	if got, want := violation.ExpectedEvents[0].ID, domain.EventID("payment_approved"); got != want {
		t.Errorf("ExpectedEvents[0].ID = %q, want %q", got, want)
	}
	if got, want := violation.ExpectedEvents[1].ID, domain.EventID("payment_rejected"); got != want {
		t.Errorf("ExpectedEvents[1].ID = %q, want %q", got, want)
	}
	if got, want := len(violation.Path), 1; got != want {
		t.Fatalf("len(Path) = %d, want %d", got, want)
	}
	if got, want := violation.Path[0].ID, domain.TransitionID("start-payment"); got != want {
		t.Errorf("Path[0].ID = %q, want %q", got, want)
	}
}

func TestValidate_RejectsUninitializedSpecification(t *testing.T) {
	_, err := Validate(domain.Specification{}, domain.Trace{})
	if !errors.Is(err, ErrUninitializedSpecification) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrUninitializedSpecification)
	}
}

func TestValidate_RejectsInvalidTraceCompleteness(t *testing.T) {
	_, err := Validate(paymentSpecification(t), domain.Trace{
		Completeness: domain.TraceCompleteness(255),
	})
	if !errors.Is(err, ErrInvalidTraceCompleteness) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidTraceCompleteness)
	}
}

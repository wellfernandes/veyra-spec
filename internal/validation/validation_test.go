package validation

import (
	"testing"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

func paymentSpecification(t *testing.T) domain.Specification {
	t.Helper()

	specification, err := domain.NewSpecification(domain.SpecificationDefinition{
		ID:           "payment-flow",
		Name:         "Payment Flow",
		InitialState: "pending",
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
	result := Validate(paymentSpecification(t), domain.Trace{
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_approved"},
		},
	})

	if !result.Valid() {
		t.Fatalf("Valid() = false, violation = %#v", result.Violation)
	}
	if got, want := result.FinalState, domain.StateID("approved"); got != want {
		t.Errorf("FinalState = %q, want %q", got, want)
	}
}

func TestValidate_EmptyTraceEndsAtInitialState(t *testing.T) {
	result := Validate(paymentSpecification(t), domain.Trace{})

	if !result.Valid() {
		t.Fatalf("Valid() = false, violation = %#v", result.Violation)
	}
	if got, want := result.FinalState, domain.StateID("pending"); got != want {
		t.Errorf("FinalState = %q, want %q", got, want)
	}
}

func TestValidate_ReportsFirstDivergenceWithContext(t *testing.T) {
	result := Validate(paymentSpecification(t), domain.Trace{
		ObservedEvents: []domain.ObservedEvent{
			{EventID: "payment_requested"},
			{EventID: "payment_cancelled"},
			{EventID: "payment_approved"},
		},
	})

	if result.Valid() {
		t.Fatal("Valid() = true, want a violation")
	}
	if got, want := result.FinalState, domain.StateID("processing"); got != want {
		t.Errorf("FinalState = %q, want %q", got, want)
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

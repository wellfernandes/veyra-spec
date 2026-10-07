package adapters

import (
	"reflect"
	"testing"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

func TestParseSpecification(t *testing.T) {
	data := []byte(`
id: payment-flow
name: Payment Flow
description: Behavioral model for a payment process.
initial_state: pending
final_states:
  - approved
  - rejected
states:
  - id: pending
    name: Pending
  - id: approved
    name: Approved
  - id: rejected
    name: Rejected
events:
  - id: payment_approved
    name: Payment Approved
  - id: payment_rejected
    name: Payment Rejected
transitions:
  - id: approve-payment
    from: pending
    event: payment_approved
    to: approved
  - id: reject-payment
    from: pending
    event: payment_rejected
    to: rejected
`)

	spec, err := ParseSpecification(data)
	if err != nil {
		t.Fatalf("ParseSpecification() returned unexpected error: %v", err)
	}

	if got, want := spec.ID(), domain.SpecificationID("payment-flow"); got != want {
		t.Fatalf("ID() = %q, want %q", got, want)
	}

	if got, want := spec.InitialState(), domain.StateID("pending"); got != want {
		t.Fatalf("InitialState() = %q, want %q", got, want)
	}

	if got, want := spec.FinalStates(), []domain.StateID{"approved", "rejected"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FinalStates() = %v, want %v", got, want)
	}

	if got, want := len(spec.Transitions()), 2; got != want {
		t.Fatalf("Transitions() length = %d, want %d", got, want)
	}
}

func TestParseSpecification_InvalidYAML(t *testing.T) {
	_, err := ParseSpecification([]byte("transitions: ["))
	if err == nil {
		t.Fatal("ParseSpecification() returned nil error, want parsing error")
	}
}

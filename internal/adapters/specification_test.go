package adapters

import (
	"reflect"
	"strings"
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

func TestParseSpecification_RejectsUnknownFieldsAndAliases(t *testing.T) {
	data := canonicalSpecificationYAML()
	unknownFields := []string{
		"initialState: pending",
		"finalStates: []",
		"state: []",
		"event: []",
		"transition: []",
		"unexpected: value",
	}

	for _, field := range unknownFields {
		t.Run(strings.SplitN(field, ":", 2)[0], func(t *testing.T) {
			if _, err := ParseSpecification([]byte(data + "\n" + field + "\n")); err == nil {
				t.Fatalf("ParseSpecification() accepted unknown field %q", field)
			}
		})
	}
}

func TestParseSpecification_RejectsScalarStatesAndEvents(t *testing.T) {
	data := canonicalSpecificationYAML()
	data = strings.Replace(data,
		"  - id: pending\n    name: Pending\n  - id: approved\n    name: Approved\n  - id: rejected\n    name: Rejected",
		"  - pending\n  - approved\n  - rejected", 1)
	data = strings.Replace(data,
		"  - id: payment_approved\n    name: Payment Approved\n  - id: payment_rejected\n    name: Payment Rejected",
		"  - payment_approved\n  - payment_rejected", 1)

	if _, err := ParseSpecification([]byte(data)); err == nil {
		t.Fatal("ParseSpecification() accepted scalar state and event entries")
	}
}

func TestParseSpecification_RejectsUnknownNestedFields(t *testing.T) {
	data := strings.Replace(canonicalSpecificationYAML(),
		"    name: Pending\n", "    name: Pending\n    label: pending\n", 1)
	if _, err := ParseSpecification([]byte(data)); err == nil {
		t.Fatal("ParseSpecification() accepted an unknown state field")
	}
}

func TestParseSpecification_RequiresStateAndEventNames(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "state name",
			data: strings.Replace(canonicalSpecificationYAML(), "    name: Pending\n", "", 1),
		},
		{
			name: "event name",
			data: strings.Replace(canonicalSpecificationYAML(), "    name: Payment Approved\n", "", 1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseSpecification([]byte(test.data)); err == nil {
				t.Fatal("ParseSpecification() accepted a missing name")
			}
		})
	}
}

func TestParseSpecification_RequiresTransitionIDs(t *testing.T) {
	data := strings.Replace(canonicalSpecificationYAML(),
		"  - id: approve-payment\n    from: pending\n",
		"  - from: pending\n", 1)
	if _, err := ParseSpecification([]byte(data)); err == nil {
		t.Fatal("ParseSpecification() accepted a transition without an ID")
	}
}

func TestParseSpecification_RejectsMultipleDocuments(t *testing.T) {
	data := canonicalSpecificationYAML() + "\n---\nid: second-specification\n"
	if _, err := ParseSpecification([]byte(data)); err == nil {
		t.Fatal("ParseSpecification() accepted multiple YAML documents")
	}
}

func canonicalSpecificationYAML() string {
	return `id: payment-flow
name: Payment Flow
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
`
}

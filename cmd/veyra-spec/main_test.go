package main

import (
	"encoding/json"
	"testing"

	"github.com/wellfernandes/veyra-spec/internal/domain"
	"github.com/wellfernandes/veyra-spec/internal/validation"
)

func TestFormatValidationResultJSON_ValidResult(t *testing.T) {
	result := validation.Result{
		CurrentState:     "pending",
		CompletionStatus: validation.CompletionStatusSatisfied,
	}

	payload, err := formatValidationResultJSON(result)
	if err != nil {
		t.Fatalf("formatValidationResultJSON() returned unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if got, want := out["consistent"], true; got != want {
		t.Fatalf("consistent = %v, want %v", got, want)
	}

	if got, want := out["valid"], true; got != want {
		t.Fatalf("valid = %v, want %v", got, want)
	}
}

func TestFormatValidationResultJSON_ViolationResult(t *testing.T) {
	result := validation.Result{
		CurrentState: "pending",
		Violation: &domain.Violation{
			TracePosition:  2,
			CurrentState:   "pending",
			ObservedEvent:  domain.ObservedEvent{EventID: "payment_cancelled"},
			ExpectedEvents: []domain.Event{{ID: "payment_approved"}, {ID: "payment_rejected"}},
		},
	}

	payload, err := formatValidationResultJSON(result)
	if err != nil {
		t.Fatalf("formatValidationResultJSON() returned unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if got, want := out["consistent"], false; got != want {
		t.Fatalf("consistent = %v, want %v", got, want)
	}

	if got := out["violation"]; got == nil {
		t.Fatal("violation field is nil, want it to be present")
	}
}

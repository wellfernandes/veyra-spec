package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
			Path: []domain.Transition{{
				ID: "request-payment", From: "created", Event: "payment_requested", To: "pending",
			}},
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

	violation, ok := out["violation"].(map[string]any)
	if !ok {
		t.Fatal("violation field is nil, want it to be present")
	}
	path, ok := violation["path"].([]any)
	if !ok || len(path) != 1 {
		t.Fatalf("violation path = %v, want one transition", violation["path"])
	}
	transition, ok := path[0].(map[string]any)
	if !ok || transition["event"] != "payment_requested" {
		t.Fatalf("violation path transition = %v, want event payment_requested", path[0])
	}
}

func TestPrintValidationResultToIncludesViolationPath(t *testing.T) {
	result := validation.Result{
		Violation: &domain.Violation{
			TracePosition: 2,
			CurrentState:  "pending",
			ObservedEvent: domain.ObservedEvent{EventID: "payment_cancelled"},
			Path: []domain.Transition{{
				ID: "request-payment", From: "created", Event: "payment_requested", To: "pending",
			}},
		},
	}
	var output bytes.Buffer

	printValidationResultTo(&output, result)

	if !strings.Contains(output.String(), "created --payment_requested--> pending") {
		t.Fatalf("output = %q, want the path leading to the violation", output.String())
	}
}

func TestFormatValidationResultJSON_IncludesCompletionPath(t *testing.T) {
	result := validation.Result{
		CurrentState:     "pending",
		CompletionStatus: validation.CompletionStatusFailed,
		CompletionViolation: &domain.CompletionViolation{
			CurrentState:        "pending",
			ExpectedFinalStates: []domain.StateID{"approved"},
			Path: []domain.Transition{{
				ID: "request-payment", From: "created", Event: "payment_requested", To: "pending",
			}},
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
	completion, ok := out["completion_violation"].(map[string]any)
	if !ok {
		t.Fatalf("completion_violation = %v, want object", out["completion_violation"])
	}
	path, ok := completion["path"].([]any)
	if !ok || len(path) != 1 {
		t.Fatalf("completion path = %v, want one transition", completion["path"])
	}
}

func TestPrintValidationResultToIncludesCompletionPath(t *testing.T) {
	result := validation.Result{
		CompletionViolation: &domain.CompletionViolation{
			CurrentState:        "pending",
			ExpectedFinalStates: []domain.StateID{"approved"},
			Path: []domain.Transition{{
				ID: "request-payment", From: "created", Event: "payment_requested", To: "pending",
			}},
		},
	}
	var output bytes.Buffer

	printValidationResultTo(&output, result)

	if !strings.Contains(output.String(), "created --payment_requested--> pending") {
		t.Fatalf("output = %q, want the path leading to completion violation", output.String())
	}
}

func TestRunValidationExitCode(t *testing.T) {
	specificationPath, tracePath := writeValidationInputs(t, "events:\n  - event: payment_approved\n")

	tests := []struct {
		name       string
		trace      string
		wantCode   int
		wantOutput string
	}{
		{
			name:       "valid trace",
			trace:      "events:\n  - event: payment_approved\n",
			wantCode:   0,
			wantOutput: "Valid: true",
		},
		{
			name:       "inconsistent trace",
			trace:      "events:\n  - event: payment_cancelled\n",
			wantCode:   1,
			wantOutput: "Valid: false",
		},
		{
			name:       "complete trace outside final state",
			trace:      "events: []\n",
			wantCode:   1,
			wantOutput: "Valid: false",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracePath = writeFile(t, filepath.Dir(tracePath), "trace.yaml", "completeness: complete\n"+test.trace)
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := runValidation(specificationPath, tracePath, false, &stdout, &stderr)
			if code != test.wantCode {
				t.Fatalf("runValidation() exit code = %d, want %d; stderr: %s", code, test.wantCode, stderr.String())
			}
			if !bytes.Contains(stdout.Bytes(), []byte(test.wantOutput)) {
				t.Fatalf("stdout = %q, want it to contain %q", stdout.String(), test.wantOutput)
			}
		})
	}
}

func TestRunValidationJSONInvalidTraceReturnsFailureCode(t *testing.T) {
	specificationPath, tracePath := writeValidationInputs(t, "events:\n  - event: payment_cancelled\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runValidation(specificationPath, tracePath, true, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("runValidation() exit code = %d, want 1; stderr: %s", code, stderr.String())
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v; output: %s", err, stdout.String())
	}
	if result["valid"] != false {
		t.Fatalf("valid = %v, want false", result["valid"])
	}
}

func writeValidationInputs(t *testing.T, trace string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	specification := strings.Join([]string{
		"id: payment-flow",
		"initial_state: pending",
		"final_states:",
		"  - approved",
		"states:",
		"  - id: pending",
		"    name: Pending",
		"  - id: approved",
		"    name: Approved",
		"events:",
		"  - id: payment_approved",
		"    name: Payment Approved",
		"transitions:",
		"  - id: approve-payment",
		"    from: pending",
		"    event: payment_approved",
		"    to: approved",
	}, "\n")
	return writeFile(t, dir, "specification.yaml", specification), writeFile(t, dir, "trace.yaml", "completeness: complete\n"+trace)
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
	return path
}

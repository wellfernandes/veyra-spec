package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestRunValidationExitCode(t *testing.T) {
	specificationPath, tracePath := writeValidationInputs(t, `events:
  - payment_approved
`)

	tests := []struct {
		name       string
		trace      string
		wantCode   int
		wantOutput string
	}{
		{
			name:       "valid trace",
			trace:      "events:\n  - payment_approved\n",
			wantCode:   0,
			wantOutput: "Valid: true",
		},
		{
			name:       "inconsistent trace",
			trace:      "events:\n  - payment_cancelled\n",
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
	specificationPath, tracePath := writeValidationInputs(t, "events:\n  - payment_cancelled\n")
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
	specification := `id: payment-flow
initial_state: pending
final_states:
  - approved
states:
  - id: pending
  - id: approved
events:
  - payment_approved
transitions:
  - from: pending
    event: payment_approved
    to: approved
`
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

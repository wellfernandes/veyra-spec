package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellfernandes/veyra-spec/internal/adapters"
	"github.com/wellfernandes/veyra-spec/internal/domain"
	"github.com/wellfernandes/veyra-spec/internal/validation"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Args)
		os.Exit(1)
	}

	if os.Args[1] == "validate" {
		jsonOutput := false
		args := os.Args[2:]
		if len(args) > 0 && args[0] == "--json" {
			jsonOutput = true
			args = args[1:]
		}
		if len(args) != 2 {
			printUsage(os.Args)
			os.Exit(1)
		}

		spec, err := adapters.LoadSpecification(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load specification %q: %v\n", args[0], err)
			os.Exit(1)
		}

		trace, err := adapters.LoadTrace(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load trace %q: %v\n", args[1], err)
			os.Exit(1)
		}

		result, err := validation.Validate(spec, trace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			payload, err := formatValidationResultJSON(result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to serialize validation result: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(payload)
			return
		}

		printValidationResult(result)
		return
	}

	if len(os.Args) != 2 {
		printUsage(os.Args)
		os.Exit(1)
	}

	path := os.Args[1]
	spec, err := adapters.LoadSpecification(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load specification %q: %v\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("Specification ID: %s\n", spec.ID())
	fmt.Printf("Initial state: %s\n", spec.InitialState())
	fmt.Printf("States: %d\n", len(spec.States()))
	fmt.Printf("Events: %d\n", len(spec.Events()))
	fmt.Printf("Transitions: %d\n", len(spec.Transitions()))

	finalStates := spec.FinalStates()
	if len(finalStates) > 0 {
		fmt.Printf("Final states: %s\n", strings.Join(stateIDsToStrings(finalStates), ", "))
	}
}

func printUsage(args []string) {
	name := filepath.Base(args[0])
	fmt.Fprintf(os.Stderr, "usage: %s <specification.yaml>\n", name)
	fmt.Fprintf(os.Stderr, "       %s validate [--json] <specification.yaml> <trace.yaml>\n", name)
}

func formatValidationResultJSON(result validation.Result) (string, error) {
	payload := map[string]any{
		"current_state":     string(result.CurrentState),
		"completion_status": completionStatusString(result.CompletionStatus),
		"consistent":        result.Consistent(),
		"completed":         result.Completed(),
		"valid":             result.Valid(),
	}

	if result.Violation != nil {
		payload["violation"] = map[string]any{
			"trace_position":  result.Violation.TracePosition,
			"current_state":   string(result.Violation.CurrentState),
			"observed_event":  string(result.Violation.ObservedEvent.EventID),
			"expected_events": eventIDsToStrings(result.Violation.ExpectedEvents),
		}
	}

	if result.CompletionViolation != nil {
		payload["completion_violation"] = map[string]any{
			"current_state":         string(result.CompletionViolation.CurrentState),
			"expected_final_states": stateIDsToStrings(result.CompletionViolation.ExpectedFinalStates),
		}
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func completionStatusString(status validation.CompletionStatus) string {
	switch status {
	case validation.CompletionStatusUnknown:
		return "unknown"
	case validation.CompletionStatusInProgress:
		return "in_progress"
	case validation.CompletionStatusNotEvaluated:
		return "not_evaluated"
	case validation.CompletionStatusSatisfied:
		return "satisfied"
	case validation.CompletionStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

func printValidationResult(result validation.Result) {
	fmt.Printf("Consistent: %t\n", result.Consistent())
	fmt.Printf("Completed: %t\n", result.Completed())
	fmt.Printf("Valid: %t\n", result.Valid())

	if result.Violation != nil {
		fmt.Printf("Violation at trace position %d\n", result.Violation.TracePosition)
		fmt.Printf("Current state: %s\n", result.Violation.CurrentState)
		fmt.Printf("Observed event: %s\n", result.Violation.ObservedEvent.EventID)
		fmt.Printf("Expected events: %s\n", strings.Join(eventIDsToStrings(result.Violation.ExpectedEvents), ", "))
	}

	if result.CompletionViolation != nil {
		fmt.Printf("Completion failed: current state %s is not in final states %s\n",
			result.CompletionViolation.CurrentState,
			strings.Join(stateIDsToStrings(result.CompletionViolation.ExpectedFinalStates), ", "))
	}
}

func stateIDsToStrings(values []domain.StateID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func eventIDsToStrings(values []domain.Event) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value.ID))
	}
	return result
}

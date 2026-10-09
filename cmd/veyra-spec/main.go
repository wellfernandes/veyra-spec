package main

import (
	"encoding/json"
	"fmt"
	"io"
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

		os.Exit(runValidation(args[0], args[1], jsonOutput, os.Stdout, os.Stderr))
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
			"path":            transitionPathToJSON(result.Violation.Path),
		}
	}

	if result.CompletionViolation != nil {
		payload["completion_violation"] = map[string]any{
			"current_state":         string(result.CompletionViolation.CurrentState),
			"expected_final_states": stateIDsToStrings(result.CompletionViolation.ExpectedFinalStates),
			"path":                  transitionPathToJSON(result.CompletionViolation.Path),
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

func runValidation(specificationPath, tracePath string, jsonOutput bool, stdout, stderr io.Writer) int {
	specification, err := adapters.LoadSpecification(specificationPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load specification %q: %v\n", specificationPath, err)
		return 1
	}

	trace, err := adapters.LoadTrace(tracePath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load trace %q: %v\n", tracePath, err)
		return 1
	}

	result, err := validation.Validate(specification, trace)
	if err != nil {
		fmt.Fprintf(stderr, "validation failed: %v\n", err)
		return 1
	}

	if jsonOutput {
		payload, err := formatValidationResultJSON(result)
		if err != nil {
			fmt.Fprintf(stderr, "failed to serialize validation result: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, payload)
	} else {
		printValidationResultTo(stdout, result)
	}

	if !result.Valid() {
		return 1
	}
	return 0
}

func printValidationResultTo(w io.Writer, result validation.Result) {
	fmt.Fprintf(w, "Consistent: %t\n", result.Consistent())
	fmt.Fprintf(w, "Completed: %t\n", result.Completed())
	fmt.Fprintf(w, "Valid: %t\n", result.Valid())

	if result.Violation != nil {
		fmt.Fprintf(w, "Violation at trace position %d\n", result.Violation.TracePosition)
		fmt.Fprintf(w, "Current state: %s\n", result.Violation.CurrentState)
		fmt.Fprintf(w, "Observed event: %s\n", result.Violation.ObservedEvent.EventID)
		fmt.Fprintf(w, "Expected events: %s\n", strings.Join(eventIDsToStrings(result.Violation.ExpectedEvents), ", "))
		printTransitionPath(w, result.Violation.Path)
	}

	if result.CompletionViolation != nil {
		fmt.Fprintf(w, "Completion failed: current state %s is not in final states %s\n",
			result.CompletionViolation.CurrentState,
			strings.Join(stateIDsToStrings(result.CompletionViolation.ExpectedFinalStates), ", "))
		printTransitionPath(w, result.CompletionViolation.Path)
	}
}

func transitionPathToJSON(path []domain.Transition) []map[string]string {
	result := make([]map[string]string, 0, len(path))
	for _, transition := range path {
		result = append(result, map[string]string{
			"id":    string(transition.ID),
			"from":  string(transition.From),
			"event": string(transition.Event),
			"to":    string(transition.To),
		})
	}
	return result
}

func printTransitionPath(w io.Writer, path []domain.Transition) {
	if len(path) == 0 {
		fmt.Fprintln(w, "Path: (no transitions)")
		return
	}

	fmt.Fprintln(w, "Path:")
	for _, transition := range path {
		fmt.Fprintf(w, "  %s --%s--> %s (%s)\n",
			transition.From, transition.Event, transition.To, transition.ID)
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

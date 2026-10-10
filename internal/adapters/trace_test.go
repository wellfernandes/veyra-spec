package adapters

import (
	"strings"
	"testing"
	"time"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

func TestParseTrace(t *testing.T) {
	data := []byte(`
completeness: complete
events:
  - event: payment_approved
    timestamp: "2024-01-01T10:00:00Z"
  - event: payment_rejected
    timestamp: "2024-01-01T10:00:05Z"
`)

	trace, err := ParseTrace(data)
	if err != nil {
		t.Fatalf("ParseTrace() returned unexpected error: %v", err)
	}

	if got, want := trace.Completeness, domain.TraceCompletenessComplete; got != want {
		t.Fatalf("Completeness = %v, want %v", got, want)
	}

	if got, want := len(trace.ObservedEvents), 2; got != want {
		t.Fatalf("ObservedEvents length = %d, want %d", got, want)
	}

	if got, want := trace.ObservedEvents[0].EventID, domain.EventID("payment_approved"); got != want {
		t.Fatalf("ObservedEvents[0].EventID = %q, want %q", got, want)
	}

	if got, want := trace.ObservedEvents[1].Timestamp.Format(time.RFC3339), "2024-01-01T10:00:05Z"; got != want {
		t.Fatalf("ObservedEvents[1].Timestamp = %q, want %q", got, want)
	}
}

func TestParseTrace_UnknownCompleteness(t *testing.T) {
	data := []byte("completeness: unknown\nevents:\n  - event: payment_approved\n")

	trace, err := ParseTrace(data)
	if err != nil {
		t.Fatalf("ParseTrace() returned unexpected error: %v", err)
	}

	if trace.Completeness != domain.TraceCompletenessUnknown {
		t.Fatalf("Completeness = %v, want %v", trace.Completeness, domain.TraceCompletenessUnknown)
	}
}

func TestParseTrace_RejectsUnknownFields(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "unknown root field",
			data: canonicalTraceYAML() + "source: packet-capture\n",
		},
		{
			name: "unknown event field",
			data: strings.Replace(canonicalTraceYAML(),
				"    timestamp: \"2024-01-01T10:00:00Z\"\n",
				"    timestamp: \"2024-01-01T10:00:00Z\"\n    source: tcp\n", 1),
		},
		{
			name: "id alias",
			data: strings.Replace(canonicalTraceYAML(), "  - event: payment_approved\n", "  - id: payment_approved\n", 1),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseTrace([]byte(test.data)); err == nil {
				t.Fatal("ParseTrace() accepted an unknown field or alias")
			}
		})
	}
}

func TestParseTrace_RequiresEventsSequence(t *testing.T) {
	if _, err := ParseTrace([]byte("completeness: complete\n")); err == nil {
		t.Fatal("ParseTrace() accepted a missing events field")
	}
	if _, err := ParseTrace([]byte("completeness: complete\nevents: []\n")); err != nil {
		t.Fatalf("ParseTrace() rejected an explicitly empty events sequence: %v", err)
	}
}

func TestParseTrace_RejectsScalarEventsAndInvalidCompleteness(t *testing.T) {
	data := strings.Replace(canonicalTraceYAML(),
		"  - event: payment_approved\n    timestamp: \"2024-01-01T10:00:00Z\"\n",
		"  - payment_approved\n", 1)
	if _, err := ParseTrace([]byte(data)); err == nil {
		t.Fatal("ParseTrace() accepted a scalar event entry")
	}

	data = strings.Replace(canonicalTraceYAML(), "completeness: complete", "completeness: Complete", 1)
	if _, err := ParseTrace([]byte(data)); err == nil {
		t.Fatal("ParseTrace() accepted non-canonical completeness casing")
	}
}

func TestParseTrace_RejectsMultipleDocuments(t *testing.T) {
	data := canonicalTraceYAML() + "\n---\ncompleteness: partial\nevents: []\n"
	if _, err := ParseTrace([]byte(data)); err == nil {
		t.Fatal("ParseTrace() accepted multiple YAML documents")
	}
}

func canonicalTraceYAML() string {
	return `completeness: complete
events:
  - event: payment_approved
    timestamp: "2024-01-01T10:00:00Z"
`
}

package adapters

import (
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
	data := []byte(`
completeness: unknown
events:
  - payment_approved
`)

	trace, err := ParseTrace(data)
	if err != nil {
		t.Fatalf("ParseTrace() returned unexpected error: %v", err)
	}

	if trace.Completeness != domain.TraceCompletenessUnknown {
		t.Fatalf("Completeness = %v, want %v", trace.Completeness, domain.TraceCompletenessUnknown)
	}
}

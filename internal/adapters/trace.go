package adapters

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wellfernandes/veyra-spec/internal/domain"
	"gopkg.in/yaml.v3"
)

type traceYAML struct {
	Completeness string               `yaml:"completeness"`
	Events       *[]observedEventYAML `yaml:"events"`
}

type observedEventYAML struct {
	Event     string `yaml:"event"`
	Timestamp string `yaml:"timestamp"`
}

func ParseTrace(data []byte) (domain.Trace, error) {
	var raw traceYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return domain.Trace{}, fmt.Errorf("decode trace YAML: %w", err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err != io.EOF {
		if err != nil {
			return domain.Trace{}, fmt.Errorf("decode trailing trace YAML: %w", err)
		}
		return domain.Trace{}, errors.New("trace YAML must contain exactly one document")
	}

	if raw.Events == nil {
		return domain.Trace{}, errors.New("trace must define events as a YAML sequence")
	}

	completeness, err := parseTraceCompleteness(raw.Completeness)
	if err != nil {
		return domain.Trace{}, err
	}

	events, err := parseObservedEvents(*raw.Events)
	if err != nil {
		return domain.Trace{}, err
	}

	return domain.Trace{
		ObservedEvents: events,
		Completeness:   completeness,
	}, nil
}

func LoadTrace(path string) (domain.Trace, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return domain.Trace{}, err
	}

	return ParseTrace(contents)
}

func parseTraceCompleteness(value string) (domain.TraceCompleteness, error) {
	switch value {
	case "", "unknown":
		return domain.TraceCompletenessUnknown, nil
	case "partial":
		return domain.TraceCompletenessPartial, nil
	case "complete":
		return domain.TraceCompletenessComplete, nil
	default:
		return domain.TraceCompletenessUnknown, fmt.Errorf("unsupported trace completeness %q", value)
	}
}

func parseObservedEvents(items []observedEventYAML) ([]domain.ObservedEvent, error) {
	events := make([]domain.ObservedEvent, 0, len(items))
	for i, item := range items {
		if strings.TrimSpace(item.Event) == "" {
			return nil, fmt.Errorf("observed event at index %d must define an event ID", i)
		}

		timestamp, err := parseTimestamp(item.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("observed event at index %d: %w", i, err)
		}

		events = append(events, domain.ObservedEvent{
			EventID:   domain.EventID(item.Event),
			Timestamp: timestamp,
		})
	}

	return events, nil
}

func parseTimestamp(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err == nil {
		return parsed, nil
	}

	parsed, err = time.Parse(time.RFC3339Nano, trimmed)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("invalid timestamp %q: %w", value, err)
}

package adapters

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wellfernandes/veyra-spec/internal/domain"
	yaml "gopkg.in/yaml.v3"
)

type traceYAML struct {
	Completeness string    `yaml:"completeness"`
	Events       yaml.Node `yaml:"events"`
}

func ParseTrace(data []byte) (domain.Trace, error) {
	var raw traceYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return domain.Trace{}, err
	}

	completeness, err := parseTraceCompleteness(raw.Completeness)
	if err != nil {
		return domain.Trace{}, err
	}

	events, err := parseObservedEvents(raw.Events)
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
	switch strings.ToLower(strings.TrimSpace(value)) {
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

func parseObservedEvents(node yaml.Node) ([]domain.ObservedEvent, error) {
	if node.Kind == 0 {
		return nil, nil
	}

	items, err := parseNodeList(node)
	if err != nil {
		return nil, err
	}

	events := make([]domain.ObservedEvent, 0, len(items))
	for _, item := range items {
		switch item.Kind {
		case yaml.ScalarNode:
			events = append(events, domain.ObservedEvent{EventID: domain.EventID(item.Value)})
		case yaml.MappingNode:
			var observed struct {
				Event     string `yaml:"event"`
				ID        string `yaml:"id"`
				Timestamp string `yaml:"timestamp"`
			}
			if err := item.Decode(&observed); err != nil {
				return nil, err
			}

			id := strings.TrimSpace(observed.Event)
			if id == "" {
				id = strings.TrimSpace(observed.ID)
			}
			if id == "" {
				return nil, fmt.Errorf("observed event is missing its event ID")
			}

			timestamp, err := parseTimestamp(observed.Timestamp)
			if err != nil {
				return nil, err
			}

			events = append(events, domain.ObservedEvent{
				EventID:   domain.EventID(id),
				Timestamp: timestamp,
			})
		default:
			return nil, fmt.Errorf("unsupported observed event definition: %s", item.ShortTag())
		}
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

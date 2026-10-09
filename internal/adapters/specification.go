package adapters

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wellfernandes/veyra-spec/internal/domain"
	"gopkg.in/yaml.v3"
)

type specificationYAML struct {
	ID           string           `yaml:"id"`
	Name         string           `yaml:"name"`
	Description  string           `yaml:"description"`
	InitialState string           `yaml:"initial_state"`
	FinalStates  []string         `yaml:"final_states"`
	States       []stateYAML      `yaml:"states"`
	Events       []eventYAML      `yaml:"events"`
	Transitions  []transitionYAML `yaml:"transitions"`
}

type stateYAML struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type eventYAML struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type transitionYAML struct {
	ID    string `yaml:"id"`
	From  string `yaml:"from"`
	Event string `yaml:"event"`
	To    string `yaml:"to"`
}

func ParseSpecification(data []byte) (domain.Specification, error) {
	var raw specificationYAML
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return domain.Specification{}, fmt.Errorf("decode specification YAML: %w", err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err != io.EOF {
		if err != nil {
			return domain.Specification{}, fmt.Errorf("decode trailing specification YAML: %w", err)
		}
		return domain.Specification{}, errors.New("specification YAML must contain exactly one document")
	}

	states := make([]domain.State, 0, len(raw.States))
	for i, state := range raw.States {
		if strings.TrimSpace(state.ID) == "" {
			return domain.Specification{}, fmt.Errorf("state at index %d must define an ID", i)
		}
		if strings.TrimSpace(state.Name) == "" {
			return domain.Specification{}, fmt.Errorf("state %q must define a name", state.ID)
		}
		states = append(states, domain.State{ID: domain.StateID(state.ID), Name: state.Name})
	}

	events := make([]domain.Event, 0, len(raw.Events))
	for i, event := range raw.Events {
		if strings.TrimSpace(event.ID) == "" {
			return domain.Specification{}, fmt.Errorf("event at index %d must define an ID", i)
		}
		if strings.TrimSpace(event.Name) == "" {
			return domain.Specification{}, fmt.Errorf("event %q must define a name", event.ID)
		}
		events = append(events, domain.Event{ID: domain.EventID(event.ID), Name: event.Name})
	}

	transitions := make([]domain.Transition, 0, len(raw.Transitions))
	for i, transition := range raw.Transitions {
		if strings.TrimSpace(transition.ID) == "" {
			return domain.Specification{}, fmt.Errorf("transition at index %d must define an ID", i)
		}
		transitions = append(transitions, domain.Transition{
			ID:    domain.TransitionID(transition.ID),
			From:  domain.StateID(transition.From),
			Event: domain.EventID(transition.Event),
			To:    domain.StateID(transition.To),
		})
	}

	finalStates := make([]domain.StateID, 0, len(raw.FinalStates))
	for _, finalState := range raw.FinalStates {
		finalStates = append(finalStates, domain.StateID(finalState))
	}

	def := domain.SpecificationDefinition{
		ID:           domain.SpecificationID(raw.ID),
		Name:         raw.Name,
		Description:  raw.Description,
		InitialState: domain.StateID(raw.InitialState),
		States:       states,
		FinalStates:  finalStates,
		Events:       events,
		Transitions:  transitions,
	}

	spec, err := domain.NewSpecification(def)
	if err != nil {
		return domain.Specification{}, err
	}

	return spec, nil
}

func LoadSpecification(path string) (domain.Specification, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return domain.Specification{}, err
	}

	return ParseSpecification(contents)
}

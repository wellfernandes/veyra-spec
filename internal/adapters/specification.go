package adapters

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/wellfernandes/veyra-spec/internal/domain"
)

type specificationYAML struct {
	ID            string    `yaml:"id"`
	Name          string    `yaml:"name"`
	Description   string    `yaml:"description"`
	InitialState  string    `yaml:"initial_state"`
	InitialState2 string    `yaml:"initialState"`
	FinalStates   yaml.Node `yaml:"final_states"`
	FinalStates2  yaml.Node `yaml:"finalStates"`
	States        yaml.Node `yaml:"states"`
	States2       yaml.Node `yaml:"state"`
	Events        yaml.Node `yaml:"events"`
	Events2       yaml.Node `yaml:"event"`
	Transitions   yaml.Node `yaml:"transitions"`
	Transitions2  yaml.Node `yaml:"transition"`
}

func ParseSpecification(data []byte) (domain.Specification, error) {
	var raw specificationYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return domain.Specification{}, err
	}

	initialState := raw.InitialState
	if strings.TrimSpace(initialState) == "" {
		initialState = raw.InitialState2
	}

	states, err := parseStates(raw.States, raw.States2)
	if err != nil {
		return domain.Specification{}, err
	}

	events, err := parseEvents(raw.Events, raw.Events2)
	if err != nil {
		return domain.Specification{}, err
	}

	transitions, err := parseTransitions(raw.Transitions, raw.Transitions2)
	if err != nil {
		return domain.Specification{}, err
	}

	finalStates, err := parseFinalStates(raw.FinalStates, raw.FinalStates2)
	if err != nil {
		return domain.Specification{}, err
	}

	def := domain.SpecificationDefinition{
		ID:           domain.SpecificationID(raw.ID),
		Name:         raw.Name,
		Description:  raw.Description,
		InitialState: domain.StateID(initialState),
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

func parseStates(nodes ...yaml.Node) ([]domain.State, error) {
	for _, node := range nodes {
		if node.Kind == 0 {
			continue
		}

		items, err := parseNodeList(node)
		if err != nil {
			return nil, err
		}

		states := make([]domain.State, 0, len(items))
		for _, item := range items {
			switch item.Kind {
			case yaml.ScalarNode:
				value := item.Value
				states = append(states, domain.State{ID: domain.StateID(value), Name: value})
			case yaml.MappingNode:
				var state struct {
					ID   string `yaml:"id"`
					Name string `yaml:"name"`
				}
				if err := item.Decode(&state); err != nil {
					return nil, err
				}
				if strings.TrimSpace(state.ID) == "" {
					state.ID = state.Name
				}
				if strings.TrimSpace(state.Name) == "" {
					state.Name = state.ID
				}
				states = append(states, domain.State{ID: domain.StateID(state.ID), Name: state.Name})
			default:
				return nil, fmt.Errorf("unsupported state definition: %s", item.ShortTag())
			}
		}
		return states, nil
	}

	return nil, nil
}

func parseEvents(nodes ...yaml.Node) ([]domain.Event, error) {
	for _, node := range nodes {
		if node.Kind == 0 {
			continue
		}

		items, err := parseNodeList(node)
		if err != nil {
			return nil, err
		}

		events := make([]domain.Event, 0, len(items))
		for _, item := range items {
			switch item.Kind {
			case yaml.ScalarNode:
				value := item.Value
				events = append(events, domain.Event{ID: domain.EventID(value), Name: value})
			case yaml.MappingNode:
				var event struct {
					ID   string `yaml:"id"`
					Name string `yaml:"name"`
				}
				if err := item.Decode(&event); err != nil {
					return nil, err
				}
				if strings.TrimSpace(event.ID) == "" {
					event.ID = event.Name
				}
				if strings.TrimSpace(event.Name) == "" {
					event.Name = event.ID
				}
				events = append(events, domain.Event{ID: domain.EventID(event.ID), Name: event.Name})
			default:
				return nil, fmt.Errorf("unsupported event definition: %s", item.ShortTag())
			}
		}
		return events, nil
	}

	return nil, nil
}

func parseTransitions(nodes ...yaml.Node) ([]domain.Transition, error) {
	for _, node := range nodes {
		if node.Kind == 0 {
			continue
		}

		items, err := parseNodeList(node)
		if err != nil {
			return nil, err
		}

		transitions := make([]domain.Transition, 0, len(items))
		for _, item := range items {
			if item.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("unsupported transition definition: %s", item.ShortTag())
			}

			var transition struct {
				ID    string `yaml:"id"`
				From  string `yaml:"from"`
				Event string `yaml:"event"`
				To    string `yaml:"to"`
			}
			if err := item.Decode(&transition); err != nil {
				return nil, err
			}

			if strings.TrimSpace(transition.ID) == "" {
				transition.ID = fmt.Sprintf("%s-%s-%s", transition.From, transition.Event, transition.To)
			}

			transitions = append(transitions, domain.Transition{
				ID:    domain.TransitionID(transition.ID),
				From:  domain.StateID(transition.From),
				Event: domain.EventID(transition.Event),
				To:    domain.StateID(transition.To),
			})
		}
		return transitions, nil
	}

	return nil, nil
}

func parseFinalStates(nodes ...yaml.Node) ([]domain.StateID, error) {
	for _, node := range nodes {
		if node.Kind == 0 {
			continue
		}

		items, err := parseNodeList(node)
		if err != nil {
			return nil, err
		}

		finalStates := make([]domain.StateID, 0, len(items))
		for _, item := range items {
			if item.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("unsupported final state definition: %s", item.ShortTag())
			}
			finalStates = append(finalStates, domain.StateID(item.Value))
		}
		return finalStates, nil
	}

	return nil, nil
}

func parseNodeList(node yaml.Node) ([]*yaml.Node, error) {
	if node.Kind == 0 {
		return nil, nil
	}

	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected a YAML list, received %s", node.ShortTag())
	}

	return node.Content, nil
}

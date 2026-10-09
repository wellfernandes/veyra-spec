# ADR 0002: Initial Specification Format (YAML)

## Status

Accepted

## Context

VeyraSpec needs an external format for defining behavioral Specifications.

The format should be:

* human-readable;
* easy for developers to write and review;
* expressive enough to represent states and transitions;
* straightforward to parse into the internal domain model.

The external format must not impose dependencies on the core validation engine.

## Decision

YAML is the initial external format for Specifications.

The canonical Specification document uses one YAML document with these fields:

* `id` and `initial_state` are required scalar values.
* `name` and `description` are optional scalar values.
* `states` is a sequence of mappings with required `id` and `name` fields.
* `events` is a sequence of mappings with required `id` and `name` fields.
* `transitions` is a sequence of mappings with required `id`, `from`, `event`, and `to` fields.
* `final_states` is an optional sequence of state IDs.

Field names use the spellings shown above. Singular and camelCase aliases are not accepted. Unknown fields and values with the wrong shape are rejected rather than silently ignored. The adapter decodes this format into typed input structures and then constructs the validated domain Specification.

YAML is an adapter concern. The validation engine will operate exclusively on internal domain structures and will not depend directly on YAML types or parsing libraries.

The intended flow is:

```text
YAML
  |
  v
Specification Adapter
  |
  v
Domain Specification
  |
  v
Validation Engine
```

The external representation may evolve through a future architectural decision. Other formats may be added through separate adapters without changing the validation engine.

## Consequences

### Positive

* Human-readable and easy to review.
* Well suited to hierarchical structures such as states and transitions.
* Familiar to developers, particularly in software infrastructure and configuration contexts.
* Keeps the external representation separated from the internal domain model.

### Negative

* YAML parsing introduces serialization and typing considerations in Go.
* The external representation may require validation before being converted into a valid domain Specification.
* The format may evolve as the domain model becomes more mature.

## Notes

This decision concerns the initial Specification adapter and does not establish YAML as the only supported format.

The architecture must not assume that YAML is the definitive or only Specification format.

Future formats may be supported through additional adapters without requiring changes to the validation engine.

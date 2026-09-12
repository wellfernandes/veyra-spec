# ADR 0002: Initial Specification Format (YAML)

## Status

Proposed

## Context

VeyraSpec needs an external format for defining behavioral Specifications.

The format should be:

* human-readable;
* easy for developers to write and review;
* expressive enough to represent states and transitions;
* straightforward to parse into the internal domain model.

The external format must not impose dependencies on the core validation engine.

## Decision

For the MVP, VeyraSpec will use **YAML as the primary external format for Specifications**.

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

The YAML representation may evolve during the MVP as the domain model and specification authoring experience become better understood.

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

This decision applies to the MVP only.

The architecture must not assume that YAML is the definitive or only Specification format.

Future formats may be supported through additional adapters without requiring changes to the validation engine.

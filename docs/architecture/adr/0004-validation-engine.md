# ADR 0004: Validation Engine Execution Flow

## Status

Accepted

## Context

The core capability of VeyraSpec is comparing expected behavior against observed behavior.

The validation process must determine whether a Trace can be followed through the behavioral model defined by a Specification and must provide meaningful diagnostic information when a divergence occurs.

The validation engine must remain deterministic and independent of external concerns such as files, YAML parsing, HTTP, databases, messaging systems, or observability platforms.

## Decision

The Validation Engine will be implemented as a **pure domain service**.

It will receive an already constructed `Specification` and `Trace` and will return a structured validation result.

The engine must not be responsible for loading, parsing, or obtaining data from external sources.

The conceptual validation flow is:

```text
Specification + Trace
        |
        v
Determine initial state
        |
        v
Process observed events
        |
        v
Identify valid transition
        |
        v
Determine next state
        |
        v
Continue execution
        |
        v
Detect divergence
        |
        v
Produce Validation Result
```

For each observed event, the engine will conceptually:

1. Determine the current state.
2. Read the next event from the Trace.
3. Determine whether a valid transition exists for the current state and observed event.
4. If a valid transition exists, move to the next state.
5. If no valid transition exists, create a structured `Violation`.
6. Preserve the execution context required to explain the divergence.
7. Continue or terminate according to the validation policy defined by the domain.
8. Produce a structured validation result.

## Responsibilities

### Validation Engine

The Validation Engine is responsible for:

* evaluating a Trace against a Specification;
* following valid transitions;
* determining the current behavioral state;
* identifying behavioral divergences;
* producing structured validation results;
* preserving the context necessary to diagnose violations.

### Outside the Validation Engine

The following responsibilities belong outside the domain engine:

* reading files;
* parsing YAML;
* parsing JSON;
* receiving HTTP requests;
* consuming messages;
* reading logs;
* interacting with databases;
* communicating with external services.

Those concerns belong to adapters or application-level components.

## Validation Result

The result must provide structured information rather than only a boolean.

At minimum, the model should be capable of representing:

* whether the execution is valid;
* the final state reached when applicable;
* detected violations;
* trace position of a violation;
* current state at the point of divergence;
* observed event;
* expected events or transitions;
* execution context required for diagnosis.

The exact structure of the result remains subject to domain refinement.

## Consequences

### Positive

* The core validation logic remains deterministic and testable.
* The engine has no dependency on external infrastructure.
* Different Specification and Trace adapters can reuse the same validation engine.
* Domain tests can focus directly on behavioral semantics.
* The architecture supports future CLI, API, test integration, or other interfaces without changing the validation logic.

### Negative

* External data must be converted into valid domain objects before validation.
* Application and adapter layers must coordinate loading and normalization.
* Additional validation may be required when constructing domain objects from external representations.

## MVP Scope

The MVP will validate an ordered Trace against a finite state machine-based Specification.

The initial validation model assumes sequential event processing.

Advanced concerns such as:

* parallel execution;
* asynchronous event ordering;
* temporal constraints;
* distributed execution semantics;
* correlation between multiple traces;

are outside the MVP and should only be introduced through explicit domain decisions.

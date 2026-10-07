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
7. Terminate validation at the first divergence because the Specification defines no next state for the observed event.
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

* whether the observed Events are consistent with the Specification;
* the State reached at the end of the observed Trace;
* whether completion is unknown, in progress, not evaluated, satisfied, or failed;
* event and completion violations;
* trace position of a violation;
* current state at the point of divergence;
* observed event;
* expected events or transitions;
* execution context required for diagnosis.

An incomplete or unknown Trace can be consistent without being a completed execution. Completion is evaluated only when the Trace is complete and the Specification defines accepted final States. When no final States are defined, the engine does not infer a completion requirement.

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

## Current Validation Contract

Validation processes an ordered Trace against the finite state machine defined by a Specification. When an observed event has no valid Transition from the current State, validation stops and reports that first divergence. Events after that point are not evaluated because the Specification provides no valid state from which to continue.

The result distinguishes event consistency from execution completion. It preserves the current State, event-violation context, completion status, and, when completion fails, the expected final States and valid Transition path.

This contract covers sequential FSM validation. Broader execution models require defined requirements and domain semantics; earlier MVP deferrals are not permanent exclusions.

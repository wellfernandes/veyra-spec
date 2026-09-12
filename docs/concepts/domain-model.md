# Domain Model

## Purpose

The VeyraSpec domain model represents the expected behavior of a software system and the observed execution that must be validated against it.

The domain is intentionally independent of external technologies, input formats, storage mechanisms, communication protocols, and infrastructure concerns.

The core model is based on the concept of a **finite state machine (FSM)**:

```text
State + Event -> Transition -> State
```

This model provides the foundation for deterministic behavioral validation while leaving room for future extensions.

## Core Concepts

### Specification

A `Specification` defines the expected behavior of a system.

It describes the states and transitions that are considered valid within a behavioral model.

A Specification represents **what should happen**.

See [specification.md](specification.md).

### State

A `State` represents a known behavioral condition within the model.

A state indicates where the system is expected to be within the specified behavior.

Examples:

* `pending`
* `processing`
* `approved`
* `rejected`
* `completed`

A State is a concept of the behavioral model. It does not necessarily correspond directly to a database value, application variable, or physical system state.

See [state.md](state.md).

### Event

An `Event` represents an observable occurrence that can participate in a behavioral transition.

Examples:

* `payment_requested`
* `payment_approved`
* `payment_rejected`

Events are the primary elements observed in a `Trace`.

See [event.md](event.md).

### Transition

A `Transition` defines a valid movement from one state to another in response to an event.

Conceptually:

```text
Current State + Event -> Next State
```

For example:

```text
pending + payment_approved -> approved
```

See [transition.md](transition.md).

### Action

An `Action` represents an operation performed by a system as part of its behavior.

Examples:

* `authorize_payment`
* `create_order`
* `send_notification`

The distinction between `Event` and `Action` is intentional:

* an Event represents something that happened or was observed;
* an Action represents something the system performs.

The exact relationship between Actions and Transitions is not fully defined for the MVP and may be refined as the domain evolves.

See [action.md](action.md).

### Trace

A `Trace` represents the observed execution of a system as an ordered sequence of events.

A Trace represents **what actually happened**.

The source of the Trace is intentionally outside the domain model. It may eventually originate from files, tests, logs, APIs, messaging systems, or distributed tracing systems.

External representations must be normalized into the internal Trace model before reaching the validation engine.

See [trace.md](trace.md).

### Execution

An `Execution` represents the behavioral execution being analyzed during validation.

It provides the context necessary to reason about the progression of a Trace through the Specification.

Execution context may include information such as:

* current state;
* processed events;
* current trace position;
* path followed through the model;
* detected violations.

The exact structure of `Execution` is intentionally subject to domain refinement.

### Validation

`Validation` is the process of comparing an observed `Trace` against a `Specification`.

The validation process determines whether the observed sequence of events can be followed through the behavioral model and identifies the first meaningful divergence when the execution is invalid.

Validation must produce structured information rather than only a boolean result.

### Violation

A `Violation` represents a divergence between the observed execution and the behavior permitted by the Specification.

A violation should provide enough context to understand:

* where the divergence occurred;
* which state was active;
* which event was observed;
* which events or transitions were expected;
* what path led to the divergence.

See [violation.md](violation.md).

## Domain Relationships

The core concepts can be viewed as follows:

```text
                 Specification
                      |
                      |
              defines expected
                      |
                      v
              State -> Transition -> State
                        ^
                        |
                       Event
                        |
                        |
                      Trace
                        |
                        v
                   Validation
                        |
                        v
                    Violation
```

The essential validation relationship is:

```text
Specification + Trace
          |
          v
      Validation
          |
          v
   Validation Result
          |
          +---- valid execution
          |
          +---- violations
```

## MVP Domain Boundary

For the MVP, the essential behavioral model is:

```text
Specification
    ├── State
    ├── Event
    └── Transition

Trace
    └── Event sequence

Validation
    └── compares Specification and Trace

Violation
    └── describes behavioral divergence
```

Actions, conditions, correlation identifiers, temporal constraints, parallel execution, asynchronous behavior, and distributed execution semantics are intentionally not required to define the MVP core model.

These concepts may be introduced later when their semantics are sufficiently defined.

## Design Principles

The domain model follows these principles:

1. **Domain independence**
   Core concepts must not depend on external technologies.

2. **Explicit behavior**
   Expected behavior must be represented explicitly rather than inferred from infrastructure.

3. **Deterministic validation**
   The same Specification and Trace should produce the same validation result.

4. **Source independence**
   The domain must not depend on how a Trace was collected.

5. **Incremental evolution**
   Future capabilities should extend the model without unnecessarily coupling the MVP to infrastructure concerns.

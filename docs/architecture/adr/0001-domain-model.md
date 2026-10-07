# ADR 0001: Domain Model Architecture

## Status

Accepted

## Context

VeyraSpec needs a robust internal conceptual model to represent both the expected behavior of a software system and its observed execution.

The validation engine must be able to determine valid next steps, follow behavioral paths, and identify divergences without depending on the internal state or implementation details of the system being observed.

The model must also remain independent of external technologies so that the validation engine can be reused with different trace sources and specification formats.

## Decision

VeyraSpec will use a **finite state machine (FSM) as the conceptual foundation of its behavioral model**.

The core behavioral relationship is:

```text
State + Event -> Transition -> State
```

The domain will contain the following core concepts:

* `Specification`
* `State`
* `Event`
* `Transition`
* `Trace`
* `Violation`

`Action` is also recognized as a domain concept representing an operation performed by the system. Its exact relationship with transitions and validation requires a separate domain decision before implementation.

The domain model is documented separately in [Domain Model](../../concepts/domain-model.md).

## Consequences

### Positive

* Provides a deterministic foundation for behavioral validation.
* Makes expected behavior explicit and modelable.
* Allows the validation engine to reason about valid states, transitions, and observed events.
* Keeps the conceptual model independent of infrastructure and external technologies.
* Provides a foundation that can be extended with more advanced behavioral semantics later.

### Negative

* Complex behavioral scenarios may require a large number of states and transitions.
* Distributed or highly asynchronous behavior may not map naturally to a simple FSM.
* Future extensions may require additional domain concepts such as conditions, guards, temporal constraints, or parallel execution semantics.

These limitations will be addressed only when the corresponding requirements and semantics are clearly defined.

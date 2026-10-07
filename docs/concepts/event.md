# Event

## Definition

An `Event` represents an observable occurrence that happened during an execution and can participate in a behavioral transition.

An Event describes **something that happened**, rather than an operation that the system performs.

## Purpose

Events are the primary observations recorded in a `Trace`.

During validation, an observed Event is evaluated against the current `State` to determine whether the Specification defines a valid `Transition`.

The core relationship is:

```text
Current State + Event -> Next State
```

## Examples

```text
payment_requested
payment_approved
payment_rejected
```

For example:

```text
pending + payment_approved -> approved
```

## Event vs. State

An Event and a State represent different concepts:

* `State` represents a condition in the behavioral model.
* `Event` represents an occurrence that may cause the execution to move between states.

An Event is therefore not itself a State.

## Current Model

Events are processed as an ordered sequence within a `Trace` by the current validator.

Concurrency, asynchronous ordering, timestamps, correlation identifiers, and distributed event relationships require explicit domain semantics before they are represented in the model.

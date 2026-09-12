# Specification

## Definition

A `Specification` defines the expected behavior of a system within a behavioral model.

It describes which states and transitions are valid and therefore establishes the rules against which an observed `Trace` is validated.

A Specification represents:

> **What should happen.**

## Purpose

The Specification is the behavioral source of truth used by the Validation Engine.

For the MVP, a Specification defines at least:

* an initial State;
* a set of States;
* Events;
* valid Transitions.

For example:

```yaml
name: payment-flow

initial_state: pending

states:
  - pending
  - approved
  - rejected

transitions:
  - from: pending
    event: payment_approved
    to: approved

  - from: pending
    event: payment_rejected
    to: rejected
```

The Validation Engine uses the Specification to determine whether a given `Trace` represents a valid execution according to the defined behavioral model.

## Specification vs. Trace

The distinction between Specification and Trace is fundamental:

```text
Specification
    |
    | expected behavior
    v
What should happen


Trace
    |
    | observed behavior
    v
What actually happened
```

Validation compares the two.

## MVP Scope

The MVP focuses on the following Specification concepts:

```text
Specification
├── initial State
├── States
├── Events
└── Transitions
```

Concepts such as:

* Actions;
* conditions;
* guards;
* rules;
* invariants;
* temporal constraints;
* parallel behavior;

may be introduced in future versions when their semantics are explicitly defined.

These concepts should not be assumed to be part of the MVP merely because they may eventually be useful.

# Specification

## Definition

A `Specification` defines the expected behavior of a system within a behavioral model.

It describes which states and transitions are valid and therefore establishes the rules against which an observed `Trace` is validated.

A Specification represents:

> **What should happen.**

## Purpose

The Specification is the behavioral source of truth used by the Validation Engine.

A Specification defines at least:

* an initial State;
* a set of States;
* Events;
* valid Transitions.

A Specification may also define one or more accepted final States. When it does, a complete Trace must end in one of those States to satisfy the completion requirement. A Specification without final States defines no completion requirement, which supports ongoing behaviors that have no natural terminal State.

For example:

```yaml
id: payment-flow
name: payment-flow

initial_state: pending
final_states:
  - approved
  - rejected

states:
  - pending
  - approved
  - rejected

events:
  - payment_approved
  - payment_rejected

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

## Current Behavioral Foundation

The current Specification model includes:

```text
Specification
├── initial State
├── States
├── Events
└── Transitions
```

Additional concepts such as:

* Actions;
* conditions;
* guards;
* rules;
* invariants;
* temporal constraints;
* parallel behavior;

may be introduced when their semantics are explicitly defined.

These concepts should not be introduced without requirements and explicit semantics merely because they may eventually be useful.

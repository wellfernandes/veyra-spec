# State

## Definition

A `State` represents a condition within the behavioral model of a system.

A State identifies the position of an execution within the behavior described by a `Specification`.

A State is a concept of the model. It does not necessarily correspond directly to:

* a database value;
* an application variable;
* an HTTP status;
* a process state;
* or another implementation detail.

## Purpose

States provide the context in which observed Events are evaluated.

Given a current State and an observed Event, the Specification determines whether a valid Transition exists and, when valid, which State should be reached next.

Conceptually:

```text
Current State
     +
Observed Event
     |
     v
Transition
     |
     v
Next State
```

## Examples

```text
pending
processing
approved
rejected
completed
```

## Initial State

A `Specification` defines an initial State from which validation begins.

For example:

```yaml
initial_state: pending
```

The validation process starts from this State and follows the Transitions defined by the Specification as Events are processed.

## Current Model

The current behavioral model treats a Specification as a finite set of States connected by valid Transitions.

Richer state semantics may be introduced when required by the domain and explicitly defined.

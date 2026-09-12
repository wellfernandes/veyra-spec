# Transition

## Definition

A `Transition` defines a valid movement from one `State` to another in response to an `Event`.

For the MVP, the fundamental model is:

```text
Current State + Event -> Next State
```

A Transition therefore defines:

* the State from which the transition can occur;
* the Event that enables the transition;
* the State reached after the transition.

## Example

Given:

```text
Current State:
pending

Event:
payment_approved

Next State:
approved
```

The Transition is:

```text
pending + payment_approved -> approved
```

## Role in Validation

During validation, the engine evaluates each observed Event against the current State.

If the Specification defines a matching Transition, the execution can move to the Transition's target State.

If no valid Transition exists, the observed execution diverges from the Specification and a `Violation` can be produced.

Conceptually:

```text
              Specification
                    |
                    v
Current State + Observed Event
                    |
             find Transition
              /           \
           found          not found
             |                |
             v                v
        Next State         Violation
```

## MVP Scope

For the MVP, a Transition is defined by:

```text
from State
+ Event
-> to State
```

Conditions, guards, Actions, temporal constraints, and other transition semantics are intentionally outside the initial model.

Those concepts may be introduced later through explicit domain decisions.

# Action

## Definition

An `Action` represents an operation performed by the system as part of its behavior.

Examples include:

* `authorize_payment`
* `send_notification`
* `create_order`

## Event vs. Action

VeyraSpec intentionally distinguishes between `Event` and `Action`:

* **Event:** something that happened or was observed.
* **Action:** something the system performs.

For example:

```text
Event:
payment_approved

Action:
send_notification
```

An Event may participate in a Transition, while an Action represents an operation performed by the system.

## Current Semantics

The distinction between Events and Actions is part of the domain vocabulary.

The exact relationship between `Action`, `Transition`, and validation is not yet defined.

Define the required Action semantics before implementing them. Their earlier deferral does not make Action behavior permanently out of scope.

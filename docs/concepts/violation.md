# Violation

## Definition

A `Violation` represents a divergence between the observed behavior described by a `Trace` and the behavior permitted by a `Specification`.

A Violation means that the observed execution cannot continue according to the behavioral model at a particular point.

## Purpose

A Violation provides diagnostic information rather than simply reporting whether validation passed or failed.

The goal is to make the divergence understandable and actionable.

A Violation should provide enough context to answer questions such as:

* Where did the divergence occur?
* Which State was active?
* Which Event was observed?
* Which Events or Transitions were expected?
* What path led to the divergence?

## Example

Suppose the Specification defines:

```text
pending + payment_approved -> approved
pending + payment_rejected -> rejected
```

and the Trace contains:

```text
payment_cancelled
```

while the current State is:

```text
pending
```

There is no valid Transition for:

```text
pending + payment_cancelled
```

The result should identify this as a Violation.

Conceptually:

```text
Trace position: 1
Current state: pending
Observed event: payment_cancelled

Expected events:
- payment_approved
- payment_rejected
```

## First Divergence

When validating a sequential Trace, the first point at which the observed execution can no longer follow the Specification is particularly important.

Validation stops at this first divergence because the Specification defines no valid next State for the observed Event. Events after the divergence are not evaluated.

The Validation Engine should preserve the context necessary to identify this first meaningful divergence. Trace positions are one-based.

This allows VeyraSpec to explain not only that the execution is invalid, but **where and why it became invalid**.

## Completion Violation

When a Trace is marked complete and the Specification defines accepted final States, ending in any other State is a completion violation. This differs from an event violation: every observed Event may follow a valid Transition, but the completed execution did not reach an accepted final State.

A completion violation records the State reached, the expected final States, and the valid Transition path. It has no observed Event or trace position because the divergence is detected at the end of the Trace.

## Validation Result

A Violation is part of the structured result produced by validation.

The result should not be reduced to a simple boolean such as:

```text
true / false
```

Instead, it should preserve the diagnostic information required to understand the behavioral divergence.

## Current Validation Foundation

The current validation foundation detects invalid observed Events and complete Traces that fail an explicitly defined final-State requirement.

Additional violation types may be introduced when their domain semantics are defined.

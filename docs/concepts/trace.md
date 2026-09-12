# Trace

## Definition

A `Trace` represents an ordered sequence of observable Events recorded during a system execution.

A Trace represents:

> **What actually happened.**

For the MVP, the fundamental representation is an ordered sequence:

```text
Trace = [Event1, Event2, Event3, ...]
```

## Purpose

The Trace provides the observed execution that will be evaluated against a `Specification`.

The Validation Engine processes the Events in order and follows the Transitions defined by the Specification.

For example:

```text
Specification:

pending
  |
  | payment_approved
  v
approved


Trace:

payment_approved
```

The engine can therefore determine that the observed execution follows the defined behavioral path.

## Trace Sources

A Trace may eventually originate from different sources, such as:

* files;
* automated tests;
* logs;
* APIs;
* messaging systems;
* simulators;
* distributed tracing systems.

However, the domain model does not depend on any particular source.

External representations must be converted into the internal `Trace` model before they are processed by the Validation Engine.

## Trace vs. Specification

The distinction is fundamental:

```text
Specification = expected behavior

Trace = observed behavior
```

Validation compares the two.

## MVP Scope

The MVP assumes that Events in a Trace have a meaningful order and are processed sequentially.

Concerns such as:

* concurrent events;
* asynchronous ordering;
* timestamps;
* correlation identifiers;
* distributed traces;
* multiple related executions;

are outside the initial Trace model.

They may be introduced later when their domain semantics are clearly defined.

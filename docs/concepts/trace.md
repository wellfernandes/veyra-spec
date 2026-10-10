# Trace

## Definition

A `Trace` represents an ordered sequence of observable Events recorded during a system execution.

A Trace represents:

> **What actually happened.**

The fundamental representation is an ordered sequence:

```text
Trace = [Event1, Event2, Event3, ...]
```

A Trace also records whether its observation is known to be partial, complete, or of unknown completeness. Unknown is the zero value so sources that cannot establish an observation boundary do not accidentally claim that an execution is complete.

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

## Trace Completeness

Events in a Trace have a meaningful order and are processed sequentially. A partial Trace describes an observed prefix; a complete Trace describes the complete execution being validated; unknown completeness means the source cannot tell whether more Events exist.

The initial YAML adapter represents a Trace with a required `events` sequence. Each entry is an object with a required `event` ID and an optional RFC 3339 `timestamp`. The optional `completeness` field accepts `unknown`, `partial`, or `complete`; omitting it means `unknown`. The sequence order determines processing order. Timestamps are retained as metadata and do not affect validation or reorder Events. Unknown fields, scalar event entries, aliases, and multiple YAML documents are rejected.

Validation always checks whether observed Events follow valid Transitions. For a partial Trace, reaching the end of the observed Events means the execution is still in progress. For unknown completeness, validation reports that completion cannot be determined. A complete Trace is checked against the Specification's accepted final States when they are defined.

Additional concerns such as:

* concurrent events;
* asynchronous ordering;
* temporal constraints;
* correlation identifiers;
* distributed traces;
* multiple related executions;

require explicit domain semantics before being added to the Trace model.

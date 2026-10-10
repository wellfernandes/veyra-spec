# ADR 0003: Agnostic Trace Model

## Status

Accepted

## Context

VeyraSpec may receive observed execution data from different sources and representations.

A Trace may eventually originate from files, automated tests, logs, APIs, messaging systems, distributed tracing systems, or other sources.

Coupling the validation engine to a specific source or representation would limit the applicability of VeyraSpec and make the core domain dependent on infrastructure concerns.

## Decision

VeyraSpec will use an internal, normalized `Trace` model independent of its external source or representation.

External trace data will be converted into the internal domain model through adapters before being processed by the validation engine.

The conceptual flow is:

```text
External Trace Source
        |
        v
      Adapter
        |
        v
  Domain Trace
        |
        v
Validation Engine
```

The Validation Engine will operate only on the internal `Trace` representation.

The same principle applies to `Event`: external event representations must be normalized before reaching the core domain.

The initial YAML Trace adapter accepts exactly one document with a required `events` sequence. Each event is a mapping with a required `event` ID and an optional RFC 3339 `timestamp`. The optional `completeness` field accepts `unknown`, `partial`, or `complete`; when omitted, it defaults to `unknown`. Event order in the YAML sequence is authoritative. Timestamps are retained as metadata and do not reorder events or affect validation. Unknown fields, scalar event entries, aliases, and additional YAML documents are rejected.

## Consequences

### Positive

* The Validation Engine remains independent of external technologies.
* Different trace sources can be supported without changing the core validation logic.
* Trace normalization provides a stable domain representation.
* Future metadata such as correlation identifiers can be added without coupling the engine to a specific transport or observability technology.

### Negative

* Each external trace source requires an adapter.
* Different trace sources may require normalization or transformation before they can be validated.
* The internal Trace model must be carefully designed to avoid leaking source-specific concepts into the domain.

The strict YAML shape applies only to the YAML adapter; other trace sources may define their own adapters and normalization rules while producing the same internal Trace.

## Current Trace Foundation

The current Trace foundation is an ordered sequence of observed Events with explicit completeness metadata.

Additional concerns such as:

* correlation identifiers;
* distributed traces;
* concurrent events;
* asynchronous execution;
* temporal validation constraints;
* message metadata;

require concrete requirements and domain semantics. Earlier MVP deferrals are not permanent exclusions.

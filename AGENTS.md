# AGENTS.md: Development Contract & AI Guidelines

> **STOP:** If you are an AI assistant (e.g., ChatGPT, Claude Code, Cursor, Copilot) assisting with this repository, you **MUST** read and follow this document before proposing, refactoring, or generating code.

## 1. Project Overview

VeyraSpec is a Go-based tool for defining expected software behavior, observing what actually happened, and validating whether the execution followed the specification.

The core comparison is:

```text
Specification = expected behavior
Trace         = observed behavior
Validation    = comparison
Violation     = behavioral divergence
```

The current behavioral model is based on a finite state machine:

```text
State + Event -> Transition -> Next State
```

The project direction is to build a robust, deterministic, and technology-independent behavioral specification and validation tool. The implementation is not constrained to an initial MVP feature set.

Existing references to MVP scope or deferred capabilities may be outdated. Do not treat them as hard product boundaries unless they are reaffirmed by a current requirement or architectural decision. This direction does not define unspecified feature semantics: new capabilities still require concrete requirements, domain decisions where applicable, tests, and synchronized documentation.

External technologies such as HTTP, message brokers, logs, databases, observability systems, or distributed tracing may eventually provide data to VeyraSpec, but they are **not part of the core validation engine**.

---

## 2. Rules for AI Agents

Before proposing or implementing changes, AI agents **MUST** understand the current domain model and architectural decisions.

Do not assume undocumented behavior.

Do not invent new architectural patterns, domain concepts, abstractions, or infrastructure without a clear requirement.

### Sources of authority

Different project documents have different responsibilities:

1. `docs/concepts/`

   * Defines domain concepts and terminology.
   * Answers: **"What does this concept mean?"**

2. `docs/architecture/adr/`

   * Records architectural decisions and their rationale.
   * Answers: **"Why was this decision made?"**

3. `AGENTS.md`

   * Defines development rules and constraints for humans and AI agents.
   * Answers: **"How should changes be made?"**

4. `README.md`

    * Describes the product's purpose, intended scope, and usage.
    * It is the best available overview of product direction, but its MVP wording may be stale and must not reinstate MVP-only restrictions.
    * It is not the authority for domain semantics or internal architectural decisions.

When these documents appear to conflict, stop and identify the inconsistency instead of silently choosing an interpretation.

### Before changing fundamental concepts

AI agents should review, as applicable:

1. `docs/concepts/domain-model.md`
2. The relevant concept document in `docs/concepts/`
3. Relevant ADRs in `docs/architecture/adr/`
4. Existing tests
5. The architectural principles defined in this document

Existing tests protect established behavior. Do not modify a test merely to make an incorrect implementation pass.

If the requested change intentionally changes established behavior, update the relevant domain documentation and tests together.

---

## 3. Product Scope and Architecture

VeyraSpec is intended to become a robust tool for defining expected behavior and validating observed behavior. The finite state machine remains the documented foundation, but previous MVP exclusions do not by themselves prohibit broader capabilities.

The essential flow is:

```text
Specification
      +
    Trace
      |
      v
Validation Engine
      |
      v
Validation Result
```

The Validation Engine is responsible for domain validation.

It must:

* receive a constructed `Specification`;
* receive a constructed `Trace`;
* determine the current `State`;
* evaluate observed `Event`s;
* identify valid `Transition`s;
* determine the next `State`;
* detect behavioral divergence;
* produce a structured validation result;
* preserve useful diagnostic context for violations.

The Validation Engine must remain independent of input and infrastructure concerns. It must **not**:

* read files;
* parse YAML;
* parse JSON;
* perform HTTP requests;
* consume messages;
* query databases;
* access external services;
* depend on observability platforms.

Those responsibilities belong outside the domain engine.

### Current validation foundation

The currently documented validation foundation includes:

* a defined initial `State`;
* a finite set of `State`s;
* defined `Event`s;
* defined `Transition`s;
* an ordered `Trace`;
* sequential event processing;
* deterministic validation;
* identification of meaningful divergences, including the first divergence.

These describe the current foundation, not a ceiling on the product. Distributed or asynchronous execution, concurrency, correlation, temporal constraints, richer Action semantics, integrations, persistence, and additional external formats are not automatically out of scope because earlier documents deferred them. Introduce them only in response to concrete requirements, with their domain semantics, architectural boundaries, tests, and documentation made explicit.

---

## 4. Architectural Principles

### 4.1 Domain First

The domain model must be understood, implemented, and tested before introducing external adapters or infrastructure.

The initial implementation should prioritize:

```text
Specification
State
Event
Transition
Trace
Validation
Violation
```

Do not introduce infrastructure merely because it may be useful in the future.

### 4.2 Keep the Core Independent

The core domain must not depend on:

* HTTP frameworks;
* databases;
* filesystem implementations;
* YAML or JSON libraries;
* Kafka, RabbitMQ, or other message brokers;
* Docker;
* cloud SDKs;
* observability platforms;
* external APIs;
* CLI frameworks.

Dependencies must point inward toward the domain.

Conceptually:

```text
External Sources
      |
      v
   Adapters
      |
      v
    Core
      |
      v
Domain behavior
```

The core must remain usable without any external technology.

### 4.3 Validation Engine Is a Domain Service

The Validation Engine is a domain-level service.

It operates on domain objects rather than external representations.

Correct:

```text
YAML
  |
  v
Specification Adapter
  |
  v
Specification
  |
  +------------------+
                     |
Trace Adapter        |
  |                  |
  v                  v
Trace --------> Validation Engine
```

Incorrect:

```text
Validation Engine
    |
    +--> reads YAML
    +--> opens files
    +--> queries databases
    +--> calls HTTP APIs
```

Do not couple validation logic to input or infrastructure concerns.

### 4.4 Adapters Translate, They Do Not Define Domain Behavior

Adapters are responsible for converting external representations into domain structures.

For example:

```text
YAML Specification -> Specification
JSON Trace         -> Trace
Log records        -> Trace
HTTP observations  -> Trace
```

Adapters must not redefine the semantics of the domain.

The domain model determines what a `Specification`, `Trace`, `State`, `Event`, and `Transition` mean.

### 4.5 Avoid Premature Complexity

Do not implement infrastructure or abstractions because they might become useful later.

In particular, do not introduce:

* distributed tracing infrastructure;
* event correlation systems;
* message brokers;
* persistence layers;
* plugin systems;
* dependency injection frameworks;
* complex event buses;
* unnecessary interfaces;
* generalized rule engines.

Only introduce additional complexity when a concrete requirement justifies it.

### 4.6 Prefer Explicit Behavior

Behavior must be:

* explicit;
* deterministic;
* readable;
* testable.

Avoid abstractions that hide the control flow without providing meaningful value.

Prefer straightforward domain logic over speculative generalization.

### 4.7 Tests Are Part of the Design

Tests should describe and protect observable domain behavior.

Prefer tests that verify:

* valid transitions;
* invalid transitions;
* state progression;
* trace validation;
* first divergence;
* violation diagnostics;
* complete validation results.

Avoid coupling tests unnecessarily to internal implementation details.

---

## 5. Domain Terminology

The following terminology must remain consistent throughout the codebase, including Go types, variables, comments, tests, and documentation.

### Specification

The formal definition of the expected behavior of a system.

A Specification defines the behavioral model used during validation.

It includes at least:

* an initial `State`;
* `State`s;
* `Event`s;
* valid `Transition`s.

### State

A condition within the behavioral model.

A State represents the current position of an execution within the specified behavior.

A State does not necessarily correspond to a database value, application variable, HTTP status, or other implementation detail.

### Event

An observable occurrence that happened during an execution and can participate in a behavioral transition.

An Event represents something that happened.

### Action

An operation performed by the system.

The distinction between `Event` and `Action` is part of the domain vocabulary:

```text
Event  = something that happened
Action = something the system performs
```

The exact relationship between `Action`, `Transition`, and validation is not fully defined. Establish the required semantics before implementing Action behavior.

Do not introduce complex Action semantics without an explicit domain decision.

### Transition

A valid movement from one `State` to another in response to an `Event`.

The current transition foundation is:

```text
Current State + Event -> Next State
```

A Transition defines:

* the source State;
* the Event;
* the target State.

### Trace

An ordered sequence of observed `Event`s recorded during an execution.

The current trace foundation is:

```text
Trace = [Event1, Event2, Event3, ...]
```

The Trace must be independent of its original source.

### Validation

The process of comparing an observed `Trace` against a `Specification`.

Validation determines whether the observed behavior can follow the transitions defined by the specification.

### Violation

A divergence between the observed behavior represented by a `Trace` and the behavior permitted by a `Specification`.

A Violation should provide enough diagnostic context to understand where and why the execution diverged.

When applicable, this includes:

* trace position;
* current State;
* observed Event;
* expected Events or Transitions;
* path leading to the divergence.

### Execution

An execution refers conceptually to the behavior represented by a Trace.

`Execution` is not currently required to be a first-class domain type.

Do not introduce an `Execution` abstraction unless a concrete requirement establishes a need for it.

### Model

The behavioral model represented by a `Specification`.

`Model` should not be used as a vague synonym for unrelated technical structures.

Prefer the more precise domain term when one exists.

---

## 6. Coding Standards & Tooling

### Language

Go.

Use idiomatic Go and follow the conventions established by the project.

### Formatting

All Go code must be formatted with:

```bash
gofmt
```

### Quality Checks

Changes should pass, as applicable:

```bash
go test ./...
go vet ./...
```

Use additional static analysis when configured by the project.

Do not introduce a new tooling requirement without justification.

### Error Handling

Errors must be handled explicitly.

Do not:

* silently ignore errors;
* use `panic` for normal application or domain control flow;
* hide meaningful errors behind generic messages.

`panic` may be appropriate only for genuinely unrecoverable programmer or initialization conditions when justified.

### Interfaces

Keep interfaces:

* small;
* focused;
* purpose-specific.

Do not create interfaces merely because a type might have multiple implementations in the future.

Define interfaces where they provide a real boundary or useful decoupling.

### Naming

Use the established domain terminology in Go identifiers.

Prefer:

```go
Specification
State
Event
Transition
Trace
Violation
```

over ambiguous or generic names when referring to domain concepts.

---

## 7. Changes to the Domain Model

Changes to fundamental domain concepts require additional care.

Examples include:

* changing the meaning of `State`;
* changing how `Event` participates in validation;
* changing the semantics of `Transition`;
* introducing guards or conditions;
* introducing complex `Action` behavior;
* changing the Trace model;
* changing how violations are detected;
* introducing concurrency or distributed semantics.

Before implementing such changes:

1. Identify the affected domain concepts.
2. Review the relevant concept documents.
3. Review affected ADRs.
4. Determine whether a new ADR or an update to an existing ADR is required.
5. Update tests that define the affected behavior.
6. Implement the change only after the intended semantics are clear.

Do not encode unresolved domain decisions directly into production code.

---

## 8. Architectural Changes

If a requested change conflicts with the architecture, explicitly identify the conflict before implementing it.

For example, if asked to:

> Add a direct database call inside the Validation Engine.

Do not implement the database call directly.

Instead, explain that this violates the core independence principle and propose an architecture consistent with the project boundaries.

Architectural changes should normally follow the direction:

```text
Domain
  ^
  |
Ports / Interfaces
  ^
  |
Adapters / Infrastructure
```

However, do not introduce a port or interface unless the domain actually requires an abstraction boundary.

Architecture must serve the domain, not the other way around.

---

## 9. External Formats

The README identifies YAML as the intended primary external format for `Specification`; ADR 0002 is still marked Proposed. Confirm and record the format decision before implementing a specification adapter.

YAML is an adapter concern.

The domain must operate on an internal `Specification` representation rather than YAML structures.

Conceptually:

```text
YAML
 |
 v
Specification Adapter
 |
 v
Domain Specification
 |
 v
Validation Engine
```

Similarly, external Trace formats must be converted into the internal `Trace` representation before validation.

The addition of another external format must not require changes to the validation semantics merely because the input representation changed.

---

## 10. Documentation Requirements

Documentation is part of the architecture.

When a change modifies domain semantics, update the relevant concept documentation.

When a change records or changes an architectural decision, create or update an ADR.

Keep documentation synchronized with the implementation.

### Concept documentation

Use:

```text
docs/concepts/
```

for definitions and domain semantics.

### Architectural decisions

Use:

```text
docs/architecture/adr/
```

for significant architectural decisions and their rationale.

Do not use ADRs merely as implementation notes.

Do not use concept documents to hide architectural decisions.

---

## 11. AI Agent Behavior

AI agents working on this repository should:

* inspect the relevant existing code before modifying it;
* inspect relevant documentation before changing domain behavior;
* preserve established terminology;
* prefer the smallest change that satisfies the requirement;
* avoid speculative abstractions;
* avoid unrelated refactoring;
* explain architectural trade-offs when they matter;
* identify conflicts instead of silently resolving them;
* keep domain logic deterministic and testable.

AI agents should **not**:

* invent requirements;
* introduce infrastructure without a concrete need;
* rewrite unrelated parts of the project;
* add dependencies merely for convenience;
* assume future requirements;
* convert every possible concept into an interface;
* reject distributed-system concepts solely because older documents deferred them; require a concrete direction and explicit domain semantics before implementing them.

---

## 12. Definition of Done for Domain Changes

A domain change should generally be considered complete only when:

* the domain behavior is clearly defined;
* relevant concept documentation is updated;
* relevant ADRs are updated or created when necessary;
* automated tests cover the intended behavior;
* `gofmt` has been applied;
* `go test ./...` passes;
* `go vet ./...` passes;
* no unnecessary infrastructure or abstraction was introduced.

The implementation should leave the repository in a state where another developer—or AI agent—can understand **what the system does, why it does it, and where that behavior belongs**.

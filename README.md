# VeyraSpec

> Model-based behavioral specification and validation for software systems.

![Status: In Development](https://img.shields.io/badge/Status-In%20Development-yellow)
![Language: Go](https://img.shields.io/badge/Language-Go-blue)

## What is VeyraSpec?

VeyraSpec is a Go-based tool for defining and validating the expected behavior of software systems.

Its core premise is simple:

> **Specify how a system should behave, observe what actually happened, and validate whether the execution followed the specification.**

VeyraSpec represents expected behavior as a behavioral model and compares it against an observed execution trace.

The initial model is based on a finite state machine (FSM):

```text
State + Event -> Transition -> Next State
```

This provides a deterministic foundation for behavioral validation while keeping the core engine independent of external technologies and trace sources.

## Why VeyraSpec?

Traditional automated tests are essential, but they often focus on isolated behavior:

* Did a function return the expected value?
* Did an API return HTTP 200?
* Did a particular component produce the expected output?

These checks do not always explain whether an entire execution followed the expected behavioral path.

VeyraSpec focuses on questions such as:

* What state should the execution be in?
* Which event was observed?
* Which transition was expected?
* Where did the observed execution diverge from the specification?
* What was the first meaningful divergence?
* Which path led to the violation?

The goal is not to replace automated tests, but to provide an additional **behavioral validation layer** that can complement them.

## How It Works

VeyraSpec works with two fundamental concepts:

### Specification

A `Specification` describes the expected behavior of a system.

It defines the behavioral model used during validation, including its initial state, states, events, and valid transitions.

A Specification represents:

> **What should happen.**

### Trace

A `Trace` represents an ordered sequence of observable events recorded during an execution.

A Trace represents:

> **What actually happened.**

### Validation

The Validation Engine compares the observed Trace against the Specification.

Conceptually:

```text
Specification                Trace
(expected behavior)          (observed behavior)
       │                           │
       └─────────────┬─────────────┘
                     ▼
                 Validation
                     │
            ┌────────┴────────┐
            ▼                 ▼
          Valid            Violation
```

When the execution does not follow a valid transition, VeyraSpec produces a structured validation result containing diagnostic information about the divergence.

## Example

A simple payment specification could be represented as:

```yaml
id: payment-flow
name: payment-flow

initial_state: pending
final_states:
  - approved
  - rejected

states:
  - id: pending
    name: Pending
  - id: approved
    name: Approved
  - id: rejected
    name: Rejected

events:
  - id: payment_approved
    name: Payment Approved
  - id: payment_rejected
    name: Payment Rejected

transitions:
  - id: approve-payment
    from: pending
    event: payment_approved
    to: approved

  - id: reject-payment
    from: pending
    event: payment_rejected
    to: rejected
```

The model defines that a payment in the `pending` state can move to either `approved` or `rejected`, depending on the observed event.

For example:

```text
pending
   |
   | payment_approved
   v
approved
```

If the observed Trace contains:

```text
payment_approved
```

the execution follows the defined behavioral path.

If the Trace instead contains:

```text
payment_cancelled
```

while the current state is `pending`, there is no valid transition for that event.

VeyraSpec can report a structured violation such as:

```text
Violation

Current state: pending
Observed event: payment_cancelled

Expected events:
- payment_approved
- payment_rejected

Trace position: 1
```

The purpose is not simply to report:

```text
FAIL
```

but to explain **where and why the execution diverged from the expected behavior**.

## Core Domain Model

VeyraSpec's behavioral model is built around:

```text
Specification
    │
    ├── Initial State
    ├── States
    ├── Events
    └── Transitions
             │
             ▼
          Validation
             ▲
             │
           Trace
             │
             ▼
         Violation
```

The fundamental relationship is:

```text
State + Event -> Transition -> State
```

`Action` is also part of the domain vocabulary, representing an operation performed by the system. Its exact relationship with transitions and validation is intentionally left open for future refinement.

## Architecture Principles

VeyraSpec is being developed with a domain-first approach.

The validation engine should remain independent of:

* HTTP frameworks;
* databases;
* message brokers;
* filesystem access;
* YAML parsing;
* observability platforms;
* external services;
* infrastructure concerns.

External representations are converted into the internal domain model through adapters.

For example:

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

Similarly, trace sources are normalized before reaching the validation engine:

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

This keeps the core validation logic deterministic, testable, and independent of the technology used to obtain the data.

## Command-Line Validation

Trace YAML requires an ordered `events` sequence. Each entry has a required `event` ID and may include a timestamp; `completeness` is optional and defaults to `unknown`. Timestamps are retained as metadata and do not affect validation order. See `examples/payment-trace-valid.yaml` for a complete example.

Validate a Trace against a Specification with:

```bash
go run ./cmd/veyra-spec validate examples/payment-spec.yaml examples/payment-trace-valid.yaml
```

Add `--json` after `validate` for machine-readable output:

```bash
go run ./cmd/veyra-spec validate --json examples/payment-spec.yaml examples/payment-trace-valid.yaml
```

The command exits with status `0` when the Trace is consistent with the Specification. It exits with a nonzero status when the Trace violates a transition, fails a defined completion requirement, or cannot be loaded or validated. A consistent partial or unknown-completeness Trace is not treated as a failure merely because it is not complete.

When validation fails, the report includes the path of valid Transitions leading to the event divergence or completion failure. JSON output represents each path as an ordered array of Transitions.

## Documentation

The project documentation is divided into two complementary areas:

```text
                         DOCUMENTATION
                              │
               ┌──────────────┴──────────────┐
               │                             │
          CONCEPTS                           ADRs
               │                             │
         "What is it?"               "Why did we decide?"
               │                             │
               ▼                             ▼
        domain-model.md              0001-domain-model.md
        state.md                     0002-specification-format.md
        event.md                     0003-trace-model.md
        transition.md                0004-validation-engine.md
        trace.md
        violation.md
        specification.md
        action.md
```

> **Note:** The ADR list above represents the architectural decisions currently documented in the repository.

### Concepts

The `docs/concepts/` directory describes the meaning and semantics of the domain concepts.

| Document           | Purpose                                                         |
| ------------------ | --------------------------------------------------------------- |
| `domain-model.md`  | Overview of the domain model and relationships between concepts |
| `specification.md` | Defines the expected behavioral model                           |
| `state.md`         | Defines behavioral states                                       |
| `event.md`         | Defines observable events                                       |
| `transition.md`    | Defines valid state transitions                                 |
| `trace.md`         | Defines observed execution traces                               |
| `violation.md`     | Defines behavioral divergences                                  |
| `action.md`        | Defines system actions and their distinction from events        |

### Architecture Decision Records

The `docs/architecture/adr/` directory records important architectural decisions and their rationale.

| Document                       | Decision                                              |
| ------------------------------ | ----------------------------------------------------- |
| `0001-domain-model.md`         | FSM as the conceptual foundation of the domain        |
| `0002-specification-format.md` | YAML as the initial Specification format              |
| `0003-trace-model.md`          | Source-independent internal Trace model               |
| `0004-validation-engine.md`    | Validation Engine responsibilities and execution flow |

The distinction is intentional:

> **Concepts explain what the system means. ADRs explain why architectural decisions were made.**

## Validation Model

The Validation Engine checks each observed Event against the current State and follows matching Transitions. It reports the first event that cannot follow the Specification, including the current State, expected Events, trace position, and valid path leading to the divergence.

Trace completeness is explicit. A partial Trace can be consistent while the execution remains in progress. Unknown completeness means the source cannot tell whether more Events exist. A complete Trace is checked against accepted final States when the Specification defines them; otherwise, the engine does not infer a completion requirement.

The core validation engine operates on domain objects and remains independent of external formats and infrastructure. Adapters convert external Specifications and trace sources into those domain objects.

## Project Status

VeyraSpec is currently **in development**.

The project is focused on establishing a well-defined domain model and a deterministic validation engine before introducing additional infrastructure or integrations.

The architecture and domain concepts may evolve as the project is implemented and tested.

## Project Philosophy

VeyraSpec follows a few core principles:

* **Domain first** — define the behavioral model before introducing infrastructure.
* **Explicit behavior** — expected behavior should be represented explicitly.
* **Deterministic validation** — the same Specification and Trace should produce the same result.
* **Technology independence** — the core should not depend on external technologies.
* **Useful diagnostics** — validation should explain violations, not only report failure.
* **Incremental evolution** — complexity should be introduced only when the domain requires it.

## License

This project is open source. See [LICENSE](LICENSE) for details.

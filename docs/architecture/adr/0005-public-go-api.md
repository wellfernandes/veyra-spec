# ADR 0005: Public Go API

## Status

Accepted

## Context

VeyraSpec currently exposes its CLI, while the domain, validation engine, and adapters live under `internal/` and therefore cannot be imported by other Go modules. The project needs a small public contract for Go callers without making implementation packages public or coupling validation to YAML, files, or test frameworks.

## Decision

The module root will provide the public package `veyra`, imported as `github.com/wellfernandes/veyra-spec`.

The API will expose independent Go types for Specification definitions, validated Specifications, Traces, observed Events, validation results, and diagnostics. `Specification` will be opaque after construction. Public identifiers will use strings, timestamps will use `time.Time`, and Trace completeness will use a public enum whose zero value is unknown.

`NewSpecification` will validate the public definition by converting it to the internal domain model and calling `domain.NewSpecification`. `Validate` will convert public values to internal types, call `validation.Validate`, and map the result back to public result and diagnostic types. No exported signature, field, or error will expose an `internal/` type.

The API is Go-only. YAML parsing, file access, CLI behavior, HTTP, protocol capture, and test-framework integrations remain outside this public contract. The CLI and current adapters will not be migrated to the facade in this decision.

The public API must preserve the current domain and validation semantics. Behavioral violations are represented in the returned result; errors are reserved for invalid input or inability to perform validation. Public error categories must be inspectable with `errors.Is` or `errors.As` and must not expose internal error types.

The API is pre-`v1` until its usage has been validated by an external-module example. Exported names and semantics should remain deliberately small to limit compatibility obligations.

## Consequences

### Positive

* Other Go modules can use validation without importing private packages.
* The core remains independent from input formats and frameworks.
* Internal implementation can evolve without directly changing consumer imports.
* An external-module test can prove that the API respects Go's `internal/` boundary.

### Negative

* Public types and error behavior become compatibility commitments once adopted.
* Conversion between public and internal representations introduces code that must preserve all diagnostics and defensive-copy guarantees.
* YAML and CLI users do not gain new capabilities from this API alone.

## Validation

The API is accepted only when a consumer in a separate Go module can build and test using the root import path alone, with no imports from `internal/`. Tests must cover successful validation, event violation, completion violation, partial and unknown Trace completeness, invalid inputs, and slice mutation isolation. The root test suite and `go vet` must also pass.

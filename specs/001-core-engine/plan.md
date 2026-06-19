# Implementation Plan: Core Engine

**Branch**: `001-core-engine` | **Date**: 2026-06-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-core-engine/spec.md`

## Summary

Build the protocol-agnostic core engine that drives all load testing in htload. Define the type system (Scenario, Phase, Step, Result), implement a multi-phase virtual-user Runner, create an HTTP/1.1+2 driver, build a live ConsoleReporter, and add template interpolation for dynamic request data.

Technical approach: Go structs with zero external dependencies in the core. `pkg/engine/` is protocol-agnostic — HTTP lives in `pkg/http/`. Runner spawns VU goroutines per-phase, collects results, and delegates to a Reporter. Templates use `text/template` with a custom funcmap.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: `github.com/google/uuid` (UUID generation). All other packages use stdlib: `net/http`, `text/template`, `context`, `sync`.
**Storage**: N/A (stateless engine)
**Testing**: `go test` with race detector against `httptest.Server`
**Target Platform**: Linux, macOS, Windows (Go cross-compile)
**Project Type**: CLI + library
**Performance Goals**: Handle 10,000 concurrent VUs without excessive memory per goroutine
**Constraints**: Single binary, zero runtime dependencies, no `net/http` in `pkg/engine/`
**Scale/Scope**: Single-machine load testing; distributed testing deferred

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Single Binary | Engine compiles to one binary, no external services | ✅ PASS |
| II. Two-File Command Pattern | CLI is FR-002; engine is library code, no CLI here | ✅ PASS |
| III. Go Stdlib First | Only `google/uuid` is new; stdlib covers http/template/sync | ✅ PASS |
| IV. No Silent Error Swallowing | All errors checked or documented; panic recovery in runner | ✅ PASS |
| V. Intent Over Implementation | Types express user intent (Scenario, Phase, Step) | ✅ PASS |
| VI. Zero External Services | No cloud/db/config needed | ✅ PASS |
| VII. Deterministic Results | `--seed` accepted; UUID is non-deterministic by design unless seeded | ✅ PASS |
| VIII. Observability | ConsoleReporter uses `slog`; JSON structured output | ✅ PASS |
| IX. File Atomicity | Result serialization prepares for atomic writes | ✅ PASS |
| X. Testing Discipline | Coverage targets ≥80% on pkg/ and internal/ | ✅ PASS |
| XI. Protocol-Agnostic Core | `pkg/engine/` has no `net/http` imports | ✅ PASS |

All principles pass. No complexity tracking violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-core-engine/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── driver.md
│   └── reporter.md
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
pkg/engine/types.go          # Scenario, Phase, Step, Result, StepContext
pkg/engine/runner.go         # Runner, multi-phase VU orchestration
pkg/http/driver.go           # HTTP driver implementing Driver interface
pkg/reporter/console.go      # ConsoleReporter implementing Reporter interface
internal/template/interpolate.go  # Template funcmap
internal/metrics/metrics.go       # Percentile, mean, min, max, median
```

## Complexity Tracking

> No violations of constitutional principles. All justified dependencies pre-approved in constitution.

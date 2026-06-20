# Tasks: Core Engine

**Branch**: `001-core-engine` | **Date**: 2026-06-19
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup

- [x] T001 Add `github.com/google/uuid` dependency via `go get github.com/google/uuid`
- [x] T002 Run `go mod tidy && go mod verify` to clean module state

---

## Phase 2: Foundational Types

Blocking prerequisites for all user stories.

- [x] T003 [P] Define engine core types in `pkg/engine/types.go` (Scenario, Phase, Step, BodySource, RateProfile, Ramp, Assertion, JSONPathAssertion, ResponseTimeAssertion, Capture, Result, StepContext)
- [x] T004 [P] Define Driver interface in `pkg/engine/driver.go` (Name, Execute, Close)
- [x] T005 [P] Define Reporter interface in `pkg/engine/reporter.go` (OnScenarioStart, OnPhaseStart, OnVUStart, OnStepStart, OnStepResult, OnVUEnd, OnPhaseEnd, OnScenarioEnd)
- [x] T006 Implement `NewStepContext()` and body resolution `BodySource.Body()` in `pkg/engine/types.go`

---

## Phase 3: User Story 1 — Load Test an HTTP Endpoint

**Goal**: Run a simple load test against an HTTP endpoint and collect Results.

**Independent Test**: Define a Scenario with one Phase and one Step pointing at an `httptest.Server`. Run it. Assert at least 25 Results with status 200 are collected.

- [x] T007 [P] [US1] Implement HTTP driver in `pkg/http/driver.go` (Name, Execute with GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS, Close)
- [x] T008 [US1] Implement Runner in `pkg/engine/runner.go` (`NewRunner`, `Run` with multi-phase VU orchestration, panic recovery, fail_fast loop termination)
- [x] T010 [US1] Wire setup and teardown execution (once globally each, using same `Driver.Execute` call with fresh `StepContext`) in `pkg/engine/runner.go`
- [x] T011 [US1] Implement status-code equality assertion in HTTP driver in `pkg/http/driver.go`
- [x] T012 [US1] Write unit tests for HTTP driver against `httptest.Server` in `pkg/http/driver_test.go`
- [x] T013 [US1] Write integration test for Runner end-to-end with `httptest.Server` in `tests/runner_test.go`

---

## Phase 4: User Story 2 — Observe Live Progress

**Goal**: See per-request results and a summary while a test runs.

**Independent Test**: Run a Scenario through ConsoleReporter and assert stderr contains ✓/✗ markers and a summary line with totals.

- [x] T014 [P] [US2] Implement ConsoleReporter in `pkg/reporter/console.go` using `log/slog` with JSON output (✓/✗ markers, phase start/end, scenario summary, thread-safe output)
- [x] T015 [US2] Wire Reporter lifecycle events into Runner in `pkg/engine/runner.go` (OnScenarioStart, OnPhaseStart, OnVUStart, OnStepStart, OnStepResult, OnVUEnd, OnPhaseEnd, OnScenarioEnd)
- [x] T016 [US2] Write unit tests for ConsoleReporter output format in `pkg/reporter/console_test.go`

---

## Phase 5: User Story 3 — Interpolate Dynamic Data

**Goal**: Generate unique request data per virtual user to prevent collisions.

**Independent Test**: Provide a Step with a template body containing `{{.UUID}}`. Run 2 VUs executing it twice. Verify all 4 outgoing request bodies contain distinct UUIDs.

- [x] T017 [P] [US3] Implement template funcmap in `internal/template/interpolate.go` (UUID, RandomString, Timestamp, TimestampNano, Env, Vars, VU.ID, VU.Iteration)
- [x] T018 [US3] Integrate template interpolation into HTTP driver for path, body, and headers in `pkg/http/driver.go`
- [x] T019 [US3] Write unit tests for template interpolation in `internal/template/interpolate_test.go`: happy-path distinct values per VU AND error-path invalid template syntax produces Result.Error

---

## Phase 6: Metrics & Polish

Cross-cutting concerns and cleanup.

- [x] T020 [P] Implement metrics aggregation in `internal/metrics/metrics.go` (Percentile, Mean, Min, Max, Median for `[]time.Duration`)
- [x] T021 Run `go test -v -race ./...` and fix any race conditions
- [x] T022 Run `golangci-lint run ./...` and fix all issues
- [x] T023 Run `gofumpt -w .` and ensure zero warnings
- [x] T024 Verify coverage ≥ 80% on `pkg/engine/`, `pkg/http/`, `pkg/reporter/`, `internal/template/`, `internal/metrics/`
- [x] T025 Update `README.md` and `AGENTS.md` with any API changes discovered during implementation
- [x] T026 Verify `pkg/engine/` does not import `net/http` or any HTTP-specific package (use `go list -f '{{.Deps}}' ./pkg/engine/ | grep net/http`)

---

## Dependency Graph

```
Phase 1 (Setup)
  ├── T001 → T002

Phase 2 (Foundational)
  ├── T003 → T006, T004, T005  (parallel types)
  ├── T004 → T007, T008, T014   (interfaces stabilize)
  ├── T005 → T014, T015         (reporter interface)
  └── T006 → T008               (StepContext needed by Runner)

Phase 3 (US1)
  ├── T007 → T011 → T012        (HTTP driver → assertion → tests)
  ├── T008 → T010 → T013        (Runner → setup/teardown → integration)
  └── T011 → T013                (assertions needed for integration)

Phase 4 (US2)
  ├── T014 → T015               (ConsoleReporter → wire into Runner)
  └── T015 → T016               (wired → test)

Phase 5 (US3)
  ├── T017 → T018               (funcmap → integrate)
  └── T018 → T019               (integrated → test)

Phase 6 (Polish)
  ├── T020 (independent, parallel with any phase after T003)
  └── T021-T026 (sequential, after all implementation)
```

## Parallel Opportunities

| Tasks | Why Parallel |
|---|---|
| T003-T006 | Foundational types/interfaces are independent |
| T007, T008, T014 | HTTP driver, Runner, and ConsoleReporter can be drafted once interfaces (T004, T005) are defined |
| T017, T020 | Template funcmap and metrics are independent of user story work once types exist |

## MVP Scope

User Story 1 only (T001-T013): A Runner that can execute a simple Scenario with an HTTP driver against an endpoint and collect Results. No reporter output, no templates, no metrics. This is the "prove the core loop works" milestone.

## Implementation Strategy

1. **Draft interfaces first** (T003-T006): Get the shape right before implementation.
2. **Implement US1 end-to-end** (T007-T013): Prove the core loop works.
3. **Layer US2 and US3** (T014-T019): Add reporting and templates on top of the working loop.
4. **Add metrics and polish** (T020-T026): Finish with shared utilities and quality gates.

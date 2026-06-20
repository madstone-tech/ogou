# Feature Specification: Core Engine

**Feature Branch**: `001-core-engine`
**Created**: 2026-06-19
**Status**: Draft
**Input**: User description: "Implement the protocol-agnostic core engine, HTTP driver, console reporter, and template interpolation. Foundation for all load testing functionality."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Load Test an HTTP Endpoint (Priority: P1)

As an engineer, I want to run a simple load test against an HTTP endpoint so that I can verify it handles concurrent traffic.

**Why this priority**: Without the core engine and HTTP driver, no load testing is possible. This is the foundational capability.

**Independent Test**: Can be fully tested by defining a Scenario with one Phase and one Step pointing at an httptest.Server, running it, and asserting that Results are collected.

**Acceptance Scenarios**:

1. **Given** a Scenario with base URL `http://localhost:8080`, one Phase with 5 VUs for 5 seconds, and one Step `GET /health`, **When** the runner executes, **Then** it produces at least 25 Results with status 200 and no errors.
2. **Given** a Scenario targeting an endpoint that returns HTTP 500, **When** the runner executes, **Then** Results have `Success=false` and `Error` describing the assertion failure.

---

### User Story 2 - Observe Live Progress (Priority: P2)

As an engineer running a long test, I want to see per-request results and a summary so that I know whether the test is healthy or failing.

**Why this priority**: Visibility into test execution is essential for debugging and confidence. Without a reporter, the tool is a black box.

**Independent Test**: Can be tested independently by running a Scenario through the ConsoleReporter and asserting that stderr contains ✓/✗ markers and a summary line.

**Acceptance Scenarios**:

1. **Given** a Scenario with one successful Step and one failing Step, **When** run with ConsoleReporter, **Then** stderr contains a ✓ marker for the success and ✗ for the failure, with step names and latencies.
2. **Given** a completed Scenario, **When** the reporter prints the summary, **Then** it shows total requests, ok count, fail count, and average latency.

---

### User Story 3 - Interpolate Dynamic Data (Priority: P3)

As an engineer testing stateful APIs, I want to generate unique request data per virtual user so that tests don't collide (e.g., duplicate email addresses).

**Why this priority**: Many APIs reject duplicate data. Template interpolation enables realistic, collision-free testing without manual data preparation.

**Independent Test**: Can be tested by providing a Step with a template body containing `{{.UUID}}` and `{{.RandomString 8}}`, running it, and verifying the outgoing request body contains different values per VU.

**Acceptance Scenarios**:

1. **Given** a Step body template `"email":"user-{{.UUID}}@test.com"`, **When** 2 VUs each execute the Step twice, **Then** all 4 outgoing request bodies contain distinct UUIDs.
2. **Given** a Step path `"/orgs/{{.Vars.orgID}}"` where `orgID` was captured in setup, **When** the Step executes, **Then** the request path contains the captured value.

---

### Edge Cases

- What happens when a template contains a syntax error? → Template evaluation error surfaces as Result.Error; the VU continues unless fail_fast is set.
- What happens when a VU goroutine panics? → Recovered at runner level, logged, treated as failed result; the VU is terminated without crashing the runner.
- What happens when the HTTP driver receives a nil Body? → No panic; empty body sent.
- What happens when an assertion specifies status 200 but the server returns 404? → Result.Success=false, Error="status 404, want 200".

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The engine MUST define a `Scenario` type containing Name, BaseURL, Setup, Phases, and Teardown.
- **FR-002**: The engine MUST define a `Phase` type containing Name, Duration, Rate (constant or ramp), and Steps.
- **FR-003**: The engine MUST define a `Step` type containing Name, Method, Path, Headers, Body, Assertions, Captures, and FailFast.
- **FR-004**: The engine MUST define a `Driver` interface with `Name()`, `Execute()`, and `Close()` methods; no `net/http` imports in `pkg/engine/`.
- **FR-005**: The `Runner` MUST execute global setup **once**, then each phase sequentially with parallel VU goroutines, then global teardown **once**.
- **FR-006**: Each VU MUST have an isolated `StepContext` with its own variable map, primed with variables captured during global setup.
- **FR-007**: The HTTP driver MUST support GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS.
- **FR-008**: The HTTP driver MUST handle request bodies from inline strings or file references, and handle nil body safely.
- **FR-009**: The HTTP driver MUST support status-code equality assertions for the v1 assertion engine.
- **FR-010**: The ConsoleReporter MUST print phase start/end, per-request markers (✓/✗), and a scenario summary to stderr.
- **FR-011**: The template engine MUST support `{{.VU.ID}}`, `{{.VU.Iteration}}`, `{{.Vars.NAME}}`, `{{.Env.NAME}}`, `{{.UUID}}`, `{{.RandomString N}}`, `{{.Timestamp}}`, `{{.TimestampNano}}`.
- **FR-012**: Template errors MUST surface as `Result.Error` without crashing the runner.
- **FR-013**: The metrics package MUST calculate min, mean, median, max, and percentile from a slice of `time.Duration`.
- **FR-014**: A VU encountering a step with `FailFast=true` and a non-success Result MUST exit its loop immediately.

### Key Entities

- **Scenario**: Top-level test definition. Contains setup, phases, teardown.
- **Phase**: A stage with a specific rate and duration. Multiple phases execute sequentially.
- **Step**: A single request definition. Contains method, path, headers, body, assertions, captures.
- **StepContext**: Mutable state scoped to a VU. Holds captured variables and iteration counter.
- **Result**: Outcome of executing one Step. Contains timing, status, success flag, error string.
- **Driver**: Protocol implementation. HTTP is v1; WebSocket/gRPC reserved.
- **Reporter**: Observer of test execution events. Console is v1.

### Assumptions

- Percentile calculation uses naive sort for v1; replaced with tdigest/HDR Histogram in a future release.
- HTTP/2 is handled automatically by Go's `net/http` client (no special configuration needed).
- Cookie persistence across steps within a VU is desirable but deferred to a later feature.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Scenario with 5 VUs looping for 5 seconds against an httptest.Server completes without panic and produces at least 25 Results.
- **SC-002**: Console output uses distinct ✓/✗ Unicode markers, step names, status codes, and latencies in ≤80 columns; a person unfamiliar with the tool can distinguish success from failure at a glance.
- **SC-003**: Template interpolation produces distinct values per VU/iteration when using `{{.UUID}}` or `{{.RandomString}}`.
- **SC-004**: `pkg/engine/` compiles without importing `net/http` or any HTTP-specific package; verified by `go list`.
- **SC-005**: Test coverage on `pkg/engine/`, `pkg/http/`, `pkg/reporter/` is ≥ 80%; coverage on `internal/template/`, `internal/metrics/` is ≥ 80%.

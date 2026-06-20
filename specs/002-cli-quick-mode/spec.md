# Feature Specification: CLI Quick Mode

**Feature Branch**: `002-cli-quick-mode`
**Created**: 2026-06-20
**Status**: Draft
**Input**: User description: "Implement the command-line interface with a quick mode for ad-hoc HTTP load testing. The user types `htload <url> [flags]` without writing a YAML file."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Ad-Hoc Load Test (Priority: P1)

A developer wants to quickly load-test an HTTP endpoint without writing a scenario file. They run `htload <url>` with optional flags to control concurrency, duration, and request count.

**Why this priority**: This is the primary daily-use path for the tool. Most users will never write a YAML scenario; they need instant gratification from the CLI.

**Independent Test**: Run `htload` against an `httptest.Server` with various flag combinations and assert exit code, output format, and request behavior.

**Acceptance Scenarios**:

1. **Given** a running HTTP server, **When** the user runs `htload http://localhost:8080/api -c 5 -d 2s`, **Then** the tool sends requests concurrently from 5 workers for 2 seconds and prints a summary.
2. **Given** a server that returns 200, **When** the user runs `htload http://localhost:8080 -n 10`, **Then** exactly 10 requests are sent and the tool exits 0.
3. **Given** a server that returns 404, **When** the user runs `htload http://localhost:8080/status/404`, **Then** the tool exits 1 because the default assertion (status == 200) fails.

---

### User Story 2 — POST with Body and Headers (Priority: P2)

A developer wants to load-test a POST endpoint with a JSON body and custom headers.

**Why this priority**: POST/PUT with bodies is a common use case for API load testing. Supporting inline body and headers makes the tool useful for realistic workloads.

**Independent Test**: Run `htload` with `-X POST`, `-D '{"a":1}'`, and `-H "Content-Type: application/json"` against an `httptest.Server` that echoes the request. Assert the body and headers are received correctly.

**Acceptance Scenarios**:

1. **Given** an echo server, **When** the user runs `htload http://localhost:8080/post -X POST -D '{"key":"value"}' -H "Content-Type: application/json"`, **Then** the server receives the exact body and header.
2. **Given** a file `payload.json` exists, **When** the user runs `htload http://localhost:8080/post -D @payload.json`, **Then** the file content is sent as the request body.

---

### User Story 3 — Threshold-Based Exit Codes (Priority: P2)

An SRE wants to integrate `htload` into a CI pipeline and fail the build if performance or reliability drops below a threshold.

**Why this priority**: Exit-code-based thresholds enable automated gating in CI/CD. This transforms the tool from a manual probe into an automated quality gate.

**Independent Test**: Run `htload` with `--fail-if-p99` and `--fail-if-rate` against controlled servers (fast and slow). Assert exit codes match threshold breaches.

**Acceptance Scenarios**:

1. **Given** a slow server (latency 20ms), **When** the user runs `htload ... --fail-if-p99 10ms`, **Then** the tool exits 1 because p99 exceeds 10ms.
2. **Given** a server that fails 5% of requests, **When** the user runs `htload ... --fail-if-rate 0.99`, **Then** the tool exits 1 because success rate is below 99%.
3. **Given** a healthy server, **When** the user runs `htload ... --fail-if-p99 100ms --fail-if-rate 0.95`, **Then** the tool exits 0 because both thresholds are met.

---

### Edge Cases

- **Zero workers**: `--workers 0` should be rejected as invalid (must be >= 1).
- **Negative duration**: `--duration -1s` should be rejected as invalid.
- **Both count and duration set**: Count takes precedence; duration is ignored.
- **Neither count nor duration set**: Use default duration (30s).
- **Invalid URL scheme**: `htload ftp://host` should exit 2 with a clear error.
- **Missing file for `-D @file`**: Exit 2 with "file not found".
- **Malformed headers**: `-H "bad"` (missing colon) should exit 2.
- **Timeout/connection refused**: Exit 1, counted as failure in results.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI entry point (`cmd/htload/main.go`) MUST use Cobra with `htload <url>` as the root command.
- **FR-002**: Quick mode MUST support the following flags with defaults:
  - `--workers` / `-c` (int, default 1): concurrent VUs
  - `--count` / `-n` (int, default 0): total request cap (0 = duration-based)
  - `--duration` / `-d` (duration, default 30s): test duration
  - `--method` / `-X` (string, default GET): HTTP method
  - `--header` / `-H` (string[], default []): repeatable headers
  - `--data` / `-D` (string, default ""): body (inline or `@file`)
  - `--fail-if-p99` (duration, default 0): exit 1 if p99 exceeds
  - `--fail-if-rate` (float64, default 0): exit 1 if success rate below
- **FR-003**: Quick mode arguments MUST compile internally to a single-phase `engine.Scenario` with one Step.
- **FR-004**: `-D "raw body"` MUST create `BodySource{Inline: ...}`; `-D @file.json` MUST create `BodySource{File: ...}`.
- **FR-005**: The tool MUST exit with code 0 (all success), 1 (any failure or threshold breach), or 2 (invalid args).
- **FR-006**: `--version` MUST print the version string.
- **FR-007**: Default assertion in quick mode MUST be `{Status: 200}` unless overridden.

### Key Entities

- **QuickModeFlags**: Parsed CLI flags before compilation to Scenario.
- **CompiledScenario**: A single-phase Scenario synthesized from flags.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `htload https://httpbin.org/get -c 10 -n 100` completes and exits 0.
- **SC-002**: `htload https://httpbin.org/status/404` exits 1 (default status assertion fails).
- **SC-003**: `htload ... -X POST -D '{"a":1}' -H "Content-Type: application/json"` sends correct body and header.
- **SC-004**: `-D @file.json` loads file content into the request body.
- **SC-005**: `--fail-if-p99 10ms` exits 1 when p99 latency exceeds 10ms.
- **SC-006**: `--fail-if-rate 0.99` exits 1 when success rate falls below 99%.
- **SC-007**: `--version` prints a non-empty version string.
- **SC-008**: Test coverage for `cmd/htload/` is ≥ 60%.

## Assumptions

- Cobra remains the CLI framework (already in `go.mod`).
- The HTTP driver from FR-001 is available and working.
- Template interpolation is not supported in quick mode v1.
- Auth provisioning is not supported in quick mode v1; headers only via `-H`.

## Out of Scope

- Scenario mode (`-f` flag) — covered by FR-003.
- Template interpolation in quick mode.
- Auth provisioning beyond static headers.
- JSON output formatting (`-o` flag deferred if complexity exceeds v1 scope).
- Docker packaging.

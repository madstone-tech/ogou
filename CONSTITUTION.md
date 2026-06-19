# Ogou Constitution

## Core Principles

### I. Single Binary

Every feature MUST compile into a single, self-contained Go binary with zero external runtime dependencies.

**Rationale**: Operator simplicity. A single binary deploys anywhere. No Docker requirement, no sidecars.

---

### II. Two-File Command Pattern

Every CLI command MUST be implemented across exactly two files:

- `<command>_cobra.go` — Cobra definition, flag registration, `init()`. Zero business logic.
- `<command>.go` — `Command` struct with an `Execute(ctx, args)` method. Zero Cobra imports.

CLI commands SHOULD share a single wire factory (`cmd/wire.go`) for common driver construction.

**Rationale**: Enables testing without Cobra. Commands can be reused from library callers. Flag parsing never leaks into business logic.

---

### III. Go Stdlib First

External dependencies MUST be justified. Default to stdlib packages:
- `net/http` for HTTP client
- `text/template` for templating
- `encoding/json` for JSON
- `regexp` for regex
- `sort` for sorting

Approved dependencies (justified):
- `github.com/spf13/cobra` — CLI framework (no stdlib equivalent)
- `github.com/tidwall/gjson` — jsonpath evaluation (stdlib has no jsonpath)
- `gopkg.in/yaml.v3` — YAML parsing (no stdlib equivalent)

Any new external dependency MUST be documented with: what it does, why stdlib can't, and what simpler alternative was rejected.

**Rationale**: Supply-chain minimisation. Every dependency is potential breakage.

---

### IV. No Silent Error Swallowing

Every error return MUST be handled or explicitly documented with a comment explaining why safe to ignore. `_, _ =` is forbidden without justification.

Functions such as `json.Marshal`, `template.Execute`, `io.Copy`, `resp.Body.Close` MUST have error returns checked.

**Rationale**: Silent errors in load testing produce misleading metrics ("all green" when requests actually failed). Explicit handling makes results trustworthy.

---

### V. Intent Over Implementation

The user-facing CLI MUST express intent, not wiring.

| Intent | Wrong | Right |
|---|---|---|
| "Test this URL" | `--vegeta-workers 10 --vegeta-duration 60s` | `htload https://api.example.com -c 10 -d 60s` |
| "Smoke test with auth" | `--auth-header Authorization --auth-value ...` | `htload -f scenario.yaml` |
| "Stop on first failure" | `--vegeta-stop-on-failure` | `fail_fast: true` in step |

**Rationale**: Abstraction matches user mental model. Users think "what am I testing?" not "how does the engine work?"

---

### VI. Zero External Services (Default)

The core engine and CLI MUST operate without any external service dependencies. No cloud accounts, no databases, no config servers.

External service reporters (Prometheus, S3, Slack) MUST be opt-in plugins, not core requirements.

**Rationale**: A load tester should work offline. Any external dependency defeats the purpose of testing infrastructure in isolation.

---

### VII. Deterministic Results

Given the same scenario YAML and the same `--seed` (optional), the tool MUST produce the same request sequence, the same template interpolations, and the same virtual user IDs.

Random generators used for `{{.UUID}}`, `{{.RandomString}}` MUST accept an optional seed. Default seed is time-based.

**Rationale**: Reproducible failures are debuggable failures. "Can't reproduce" is the worst bug report.

---

### VIII. Observability

- All structured logging MUST use `log/slog` with JSON output.
- `fmt.Println` and `fmt.Printf` are forbidden in production code paths.
- Console reporter output goes to stderr; JSON reports go to stdout or file.

**Rationale**: Machine-parseable output enables piping to `jq`, ingestion into monitoring systems, and CI artifact processing.

---

### IX. File Atomicity

All report file writes MUST follow temp-then-rename:

1. Write to `.tmp`-suffixed path in same directory as target.
2. `fsync` the file descriptor.
3. `os.Rename` to final target.

Interrupted writes leave a partial `.tmp`; the original result file (if any) remains intact.

**Rationale**: Prevents partial/corrupt JSON reports from being ingested by downstream tools.

---

### X. Testing Discipline

- Coverage on `pkg/engine/`, `pkg/http/`, `pkg/scenario/`, and `pkg/auth/` MUST be ≥ 80%.
- Coverage on `cmd/` MUST be ≥ 60%.
- Round-trip property tests MUST exist for YAML parser ↔ struct ↔ YAML.
- No new functionality MAY be merged without tests.
- Every development session MUST end with `go test ./...` passing.

**Rationale**: Load testing tools must themselves be tested. A buggy assertion engine reports false positives, destroying trust in results.

---

### XI. Protocol-Agnostic Core

The `pkg/engine/` types and `Runner` MUST NOT import `net/http` or any HTTP-specific package. The engine is protocol-agnostic.

HTTP is the v1 driver. The `Driver` interface MUST permit future WebSocket and gRPC drivers without engine changes.

**Rationale**: Architectural investment now prevents a rewrite later. The core is the invariant; drivers are replaceable.

---

## Additional Constraints

### CLI Design

- Quick mode: `htload <url>` is the primary invocation. Everything else is optional.
- Scenario mode: `htload -f <file>` is the secondary invocation.
- Exit codes are part of the contract (see SPEC.md §6). CI depends on them.
- `--version` MUST print module version from `go.mod` or build-time ldflags.
- `--dry-run` MUST be supported for scenario mode: parse, validate, interpolate templates once, print what would be sent, exit 0 without sending.

### Result Integrity

- Every `Result` struct MUST be serialisable to JSON without data loss.
- Timestamps MUST be UTC ISO 8601.
- Durations MUST be nanosecond integers, not Go duration strings.
- Latency distributions MUST use HDR Histogram (adapted) or tdigest for accurate percentiles at scale, not naive sorting.

### Code Style

- Formatting: `gofumpt`. MUST pass with zero warnings before any commit.
- Linting: `golangci-lint`. MUST pass with zero warnings before any commit.
- Both tools MUST run as pre-commit gates.

---

## Development Workflow & Quality Gates

### Feature Lifecycle (Speckit)

1. Feature spec created in `specs/<###-feature-name>/spec.md`.
2. Implementation plan created in `specs/<###-feature-name>/plan.md`.
3. **Constitution Check** — plan gated against all 11 principles before Phase 0 research begins.
4. Tasks generated in `specs/<###-feature-name>/tasks.md` organised by phase (P0 → P1 → P2 → P3).
5. Implementation follows two-file command pattern (Principle II).
6. All CI Workflow Gate steps MUST pass before any merge.
7. `gofumpt` and `golangci-lint` MUST pass with zero warnings (enforced by CI).

### Complexity Justification

Any violation of the protocol-agnostic rule, any new external dependency, or any design choice deviating from Principle III (Go Stdlib First) MUST be documented in the plan's Complexity Tracking table with:
- The specific violation or deviation.
- Why it is necessary.
- What simpler alternative was considered and why it was rejected.

### CI Workflow Gate

Every implementation MUST pass all steps defined in `.github/workflows/ci.yml` before being considered complete.

The CI pipeline enforces:

1. `go mod tidy` — no uncommitted module changes.
2. `go mod verify` — all module checksums valid.
3. `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...`
   — all tests pass with race detector enabled.
4. `go build -v ./cmd/htload` — binary compiles cleanly.
5. `golangci-lint` — zero issues.
6. Constitution Check — no unprincipled deviations.

### Session Close Checklist

Every development session MUST end with all CI Workflow Gate steps passing locally:

- [ ] `go mod tidy && go mod verify` — clean.
- [ ] `go test -v -race ./...` — all tests pass, no data races.
- [ ] Coverage ≥ 80% on changed packages in `pkg/engine/`, `pkg/http/`, `pkg/scenario/`, `pkg/auth/`.
- [ ] Coverage ≥ 60% on changed packages in `cmd/`.
- [ ] `go build -v ./cmd/htload` — binary compiles.
- [ ] `golangci-lint run ./...` — zero issues.
- [ ] All new errors handled or explicitly justified (Principle IV).

---

## Governance

This constitution is the highest-authority document for Ogou (htload) development.

### Amendment Procedure

1. Propose amendment in writing, referencing the principle number and version.
2. State the technical rationale.
3. Increment version per Semantic Versioning.
4. Update `LAST_AMENDED_DATE`.

### Compliance Review

- All PRs MUST be reviewed against applicable principles.
- Security-relevant changes (Principles IV, VII, IX, X) require explicit sign-off.
- The constitution MUST be reviewed at each major release (v0.1.0, v1.0.0).

**Version**: 1.0.0 | **Ratified**: 2026-06-19 | **Last Amended**: 2026-06-19

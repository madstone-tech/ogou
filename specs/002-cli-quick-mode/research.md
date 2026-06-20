# Research: CLI Quick Mode

**Date**: 2026-06-20
**Feature**: CLI Quick Mode

## Research Topics

### Topic 1: Two-File Command Pattern

**Question**: How should the CLI command be structured to comply with Constitution Principle II?

**Decision**:
Split `cmd/htload/` into two files:
- `root_cobra.go` — Cobra command definition, flag registration, `init()`, `Execute()`. Zero business logic.
- `main.go` — `Command` struct with an `Execute(ctx, args)` method. Zero Cobra imports.

**Rationale**:
- Constitution Principle II mandates this pattern: Cobra concerns and business logic must not coexist in the same file.
- Enables testing the command as a library without importing Cobra. Callers can invoke `cmd.Execute(ctx, url, flags)` directly.
- Prevents flag-parsing logic from leaking into business logic, keeping the command reusable from other Go programs.

**Alternatives considered**:
- Single file mixing Cobra and logic (current state): rejected — violates Constitution II; untestable without Cobra, and the current `cmd/htload/main.go` already mixes both.
- Three files (adding `wire.go` now): rejected — wire factory is optional for v1; the two-file pattern is sufficient and avoids premature abstraction.

---

### Topic 2: Flag-to-Scenario Compilation

**Question**: How should quick-mode CLI flags be translated into an `engine.Scenario`?

**Decision**:
Compile flags into a single-phase `Scenario` with exactly one `Step`:

- `BaseURL` = scheme + host from parsed URL
- `Phase.Name` = `"load"`
- `Phase.Duration` = flag value (default 30s)
- `Phase.Rate.Constant` = flagWorkers
- `Step.Name` = `"request"`
- `Step.Method` = flag value (default GET)
- `Step.Path` = parsedURL.Path + parsedURL.RawQuery
- `Step.Headers` = parsed -H headers
- `Step.Body` = resolved BodySource
- `Step.Assertions` = `[{Status: 200}]` (default)

**Rationale**:
- Reuses the existing `engine.Runner` and `engine.Scenario` without duplicating execution logic.
- A single Phase with one Step is the minimal valid Scenario; it naturally supports the quick-mode use case.
- Using the engine’s own types guarantees that quick mode and scenario mode produce identical runtime behavior.

**Alternatives considered**:
- Dedicated quick-mode execution path bypassing `Scenario`: rejected — duplicates logic already in `Runner` and `Driver`; diverges from scenario mode.
- Multiple Phases (e.g., ramp-up + steady): rejected — over-engineering for v1; future FRs can extend the compilation rule when ramp/spike flags are added.

---

### Topic 3: Threshold Evaluation Strategy

**Question**: When and how should `--fail-if-p99` and `--fail-if-rate` be evaluated?

**Decision**:
Evaluate thresholds after `Runner.Run()` returns, computing p99 and success rate from the `[]*Result` slice.

- p99: sort latencies, pick index `ceil(0.99 * n) - 1`.
- Success rate: `successCount / totalCount`.
- If any threshold is breached, return an error that the CLI layer translates to exit code 1.

**Rationale**:
- Post-run evaluation is deterministic: all results are known, no streaming approximations needed.
- Keeps threshold logic out of the engine core (which is protocol-agnostic) and in the CLI layer where exit codes live.
- The `Result` slice already contains `Latency` and `Success`; no additional instrumentation required.

**Alternatives considered**:
- Streaming evaluation inside `Runner`: rejected — violates Constitution XI (protocol-agnostic core should not know about CLI exit codes); also complicates the runner lifecycle.
- Inside `Reporter`: rejected — reporters are event sinks, not decision makers; mixing policy into reporting breaks separation of concerns.

---

### Topic 4: Exit Code Model

**Question**: What exit codes should the CLI return?

**Decision**:

| Code | Condition |
|------|-----------|
| 0 | All requests succeeded and all active thresholds met |
| 1 | Any request failed (assertion mismatch, transport error) OR any threshold breached |
| 2 | Invalid arguments (malformed URL, missing file, bad flag value, Cobra parse error) |

**Rationale**:
- Three codes align with common CLI conventions (success / failure / usage error).
- Code 1 is the CI-critical signal: a test that finds problems must exit non-zero so CI pipelines fail.
- Code 2 separates operator mistakes (bad args) from runtime failures, making debugging faster.

**Alternatives considered**:
- Binary exit (0 or 1 only): rejected — conflates argument errors with test failures; operators cannot distinguish `htload --bad-flag` from `htload url-that-404s`.
- HTTP-specific codes (e.g., 404 → exit 4): rejected — leaks HTTP semantics into the CLI contract; the tool may support WebSocket/gRPC later.

---

### Topic 5: Body Source Strategy

**Question**: How should the `-D` flag translate into `BodySource`?

**Decision**:
- `-D "raw body"` → `BodySource{Inline: "raw body"}`
- `-D @file.json` → `BodySource{File: "file.json"}` (relative or absolute path, resolved at execution time)

**Rationale**:
- Prefix matching (`@`) is familiar from `curl -d @file` and reduces the need for an extra flag.
- `BodySource` already supports `Inline` and `File` fields; no new types needed.
- File resolution happens at execution time (inside `BodySource.Body()`), not at flag-parse time, so the file can be generated after the command is typed.

**Alternatives considered**:
- Separate `--data-file` flag: rejected — adds another flag to remember; `@` prefix is a well-established convention.
- Always treat `-D` as a file path: rejected — breaks the common case of inline JSON; users would need `echo '{"a":1}' | htload ...` which is more complex.

---

## Resolved Clarifications

| # | Original Question | Resolution |
|---|---|---|
| 1 | CLI command structure | Two-file pattern: `root_cobra.go` + `main.go` |
| 2 | Flag translation | Compile to single-phase Scenario with one Step |
| 3 | Threshold timing | Post-run evaluation from `[]*Result` |
| 4 | Exit codes | 0 = success, 1 = failure/breach, 2 = invalid args |
| 5 | Body loading | `-D inline` → Inline; `-D @file` → File |

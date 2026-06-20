# Tasks: CLI Quick Mode

**Branch**: `002-cli-quick-mode` | **Date**: 2026-06-20
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup

- [ ] T001 Verify branch `002-cli-quick-mode` is clean and buildable (`go build ./cmd/htload`)
- [ ] T002 Run existing tests and record baselines (`go test ./...`)
- [ ] T003 Verify directory structure matches plan (cmd/htload/, existing engine/)

---

## Phase 2: Foundational (Two-File Command Refactor)

Blocking prerequisites for all user stories. Refactors existing `cmd/htload/main.go` into the two-file command pattern per Constitution Principle II.

- [ ] T004 [P] Create `cmd/htload/root.go` with Cobra root command, flag registration, `init()`, and `Execute()`
- [ ] T005 Refactor `cmd/htload/main.go` to remove all Cobra imports, add `Run(url string, flags QuickModeFlags) error` entry point
- [ ] T006 Wire `root.go` `Execute()` → `main.go` `Run()` with compiled flags
- [ ] T007 Wire `root.go` to handle `--version` as a `PersistentFlag` that short-circuits before `Run()`
- [ ] T008 Wire `root.go` to preserve scenario mode path (`--file` → existing `runScenario()`)

---

## Phase 3: User Story 1 — Ad-Hoc Load Test (P1)

**Goal**: Run `htload <url> [flags]` to execute a quick load test without a YAML file.

**Independent Test**: Run `htload` against an `httptest.Server` with `-c 5 -d 2s` and assert ≥5 requests sent, exit 0 for 200, exit 1 for 404.

- [ ] T009 [US1] Implement flag-to-Scenario compilation in `cmd/htload/main.go`: parse URL → build `engine.Scenario` with single Phase + Step
- [ ] T010 [US1] Implement `QuickModeFlags` struct with all fields (URL, Workers, Count, Duration, Method, Headers, Data) in `cmd/htload/main.go`
- [ ] T011 [US1] Validate quick-mode flags in `cmd/htload/main.go`:
  - URL scheme is `http` or `https`
  - Workers >= 1
  - Duration > 0
  - Count >= 0
  (any violation returns exit code 2)
- [ ] T012 [US1] Integrate `engine.Runner` + `http.Driver` + `reporter.ConsoleReporter` into `Run()` for quick mode execution
- [ ] T013 [US1] Write unit tests in `cmd/htload/main_test.go`: flag compilation produces valid Scenario struct, base URL and path split correctly
- [ ] T014 [US1] Write integration tests in `cmd/htload/main_test.go`: `Run()` with `httptest.Server`, assert request count and exit code for 200/404

---

## Phase 4: User Story 2 — POST with Body and Headers (P2)

**Goal**: Send POST/PUT requests with inline body, file body, and custom headers.

**Independent Test**: Run `htload` with `-X POST -D '{"key":"value"}' -H "Content-Type: application/json"` against echo server. Assert body and headers match exactly.

- [ ] T015 [P] [US2] Implement header parsing in `cmd/htload/main.go`: split `-H` strings on first `:` into `map[string]string`
- [ ] T016 [P] [US2] Implement body source resolution in `cmd/htload/main.go`: `-D "raw"` → `BodySource{Inline}`, `-D @file` → `BodySource{File}`
- [ ] T017 [US2] Inject default assertion `{Status: 200}` into compiled Step in `cmd/htload/main.go`
- [ ] T018 [US2] Write unit tests in `cmd/htload/main_test.go`: header parsing edge cases (malformed, extra colons), body source resolution
- [ ] T019 [US2] Write integration tests in `cmd/htload/main_test.go`: POST with body and headers against `httptest.Server`, assert echoed request matches

---

## Phase 5: User Story 3 — Threshold-Based Exit Codes (P2)

**Goal**: After quick mode execution, evaluate `--fail-if-p99` and `--fail-if-rate` thresholds and return appropriate exit code.

**Independent Test**: Run `htload` with `--fail-if-p99 10ms` against a 20ms server and assert exit 1. Run with `--fail-if-rate 0.99` against a 95% server and assert exit 1. Run against a healthy server and assert exit 0.

- [ ] T020 [P] [US3] Implement success-rate computation in `cmd/htload/main.go`: `ok / total` from `[]*engine.Result`
- [ ] T021 [P] [US3] Implement p99 latency computation in `cmd/htload/main.go`: sort `Result.Latency`, pick index `ceil(0.99*n) - 1`
- [ ] T022 [US3] Implement threshold evaluation in `cmd/htload/main.go`: compare computed p99/rate against flags, return breach boolean + reason
- [ ] T023 [US3] Implement exit code determination in `cmd/htload/main.go`: 0=all pass, 1=failure/breach, 2=invalid args
- [ ] T024 [US3] Write unit tests in `cmd/htload/main_test.go`: success-rate and p99 calculation on synthetic `Result` slices
- [ ] T025 [US3] Write unit tests in `cmd/htload/main_test.go`: exit code determinism matrix (all combinations of argErr, runErr, thresholds, results)

---

## Phase 6: CLI Flag Parsing Tests

- [ ] T026 [P] Write `cmd/htload/root_test.go` with unit tests for Cobra flag parsing:
  - `-c 0` rejected
  - `-d -1s` rejected
  - `-H "bad"` rejected (missing colon)
  - `-D @missing.json` validated at execution time
  - `-X INVALID` accepted (HTTP driver handles invalid method)
  - `--version` prints non-empty string and short-circuits
  - positional arg required (no URL → error)

---

## Phase 7: Polish & Cross-Cutting Concerns

- [ ] T027 Wire `--file` scenario mode through `root.go` → `main.go` without regression; `runScenario()` unchanged
- [ ] T028 Verify `cmd/htload/` does not import `net/http` directly (import boundary per Constitution XI)
- [ ] T029 Run `go test ./...` and fix any failures
- [ ] T030 Run `go test -race ./...` and fix any data races
- [ ] T031 Run `golangci-lint run ./...` and fix all issues
- [ ] T032 Run `gofumpt -w .` and ensure zero warnings
- [ ] T033 Verify coverage ≥ 60% for `cmd/htload/` (`go test -coverprofile=... ./cmd/htload/`)
- [ ] T034 Update `README.md` with quick mode usage examples from `quickstart.md`
- [ ] T035 Update `AGENTS.md` if any CLI API changes discovered during implementation

---

## Dependency Graph

```
Phase 1 (Setup)
  ├── T001 → T002 → T003

Phase 2 (Refactor)
  ├── T004 → T005 → T006
  ├── T007 (version flag)
  └── T008 (scenario mode preservation)

Phase 3 (US1 — Ad-Hoc Load Test)
  ├── T009 → T010 → T011 → T012
  └── T013, T014 (tests)

Phase 4 (US2 — POST/Body/Headers)
  ├── T015, T016 (parallel)
  ├── T017 → T018, T019 (tests)

Phase 5 (US3 — Thresholds)
  ├── T020, T021 (parallel)
  ├── T022 → T023
  └── T024, T025 (tests)

Phase 6 (Root Tests)
  └── T026 (independent, parallel with Phase 3+)

Phase 7 (Polish)
  ├── T027 (regression check)
  └── T028-T035 (quality gates)
```

## Parallel Opportunities

| Tasks | Why Parallel |
|---|---|
| T004 + T005 | Cobra file and business-logic file are independent once interface agreed |
| T015 + T016 | Header parsing and body resolution are independent |
| T020 + T021 | Success-rate and p99 computation are independent |
| T026 | Flag parsing tests are independent of execution tests |

## MVP Scope

User Story 1 only (T001-T014): A CLI that runs `htload <url>` with `-c` and `-d` flags, compiles to a Scenario, executes via Runner, and exits 0/1/2. No body/headers (US2) or thresholds (US3) in the MVP.

## Implementation Strategy

1. **Refactor first** (T004-T008): Split the existing `main.go` into the two-file pattern. This is blocking.
2. **Implement US1 end-to-end** (T009-T014): Get the basic quick mode working.
3. **Layer US2 and US3** (T015-T025): Add body/headers, then thresholds.
4. **Test root parsing** (T026): Independent flag parsing validation.
5. **Polish and quality gates** (T027-T035): Regression check, coverage, lint, format.

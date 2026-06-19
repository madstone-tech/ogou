# Tasks: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE]
**Spec**: [link] | **Plan**: [link]

---

## Phase 0: Setup & Baselines

- [ ] T001: Verify branch clean and buildable
- [ ] T002: Run existing tests and record baselines
- [ ] T003: Verify directory structure matches plan

## Phase 1: Core Types

- [ ] T004: Define engine types (Scenario, Phase, Step, Result)
- [ ] T005: Define Driver and Reporter interfaces
- [ ] T006: Define StepContext and template data structures

## Phase 2: Runner Implementation

- [ ] T007: Implement basic Runner.Run() loop
- [ ] T008: Implement VU goroutine spawning
- [ ] T009: Implement phase sequencing
- [ ] T010: Implement global setup and teardown

## Phase 3: HTTP Driver

- [ ] T011: Implement HTTP driver conforming to Driver interface
- [ ] T012: Handle GET, POST, PUT, DELETE, PATCH
- [ ] T013: Handle request body from inline/file
- [ ] T014: Handle headers and assertions

## Phase 4: Reporter

- [ ] T015: Implement ConsoleReporter
- [ ] T016: Implement JSONReporter
- [ ] T017: Implement DiscardReporter

## Phase 5: Template Engine

- [ ] T018: Implement template interpolation with funcmap
- [ ] T019: Add UUID, RandomString, Timestamp functions
- [ ] T020: Add Env and Vars lookup

## Phase 6: Validation

- [ ] T021: Implement scenario validation
- [ ] T022: Implement YAML parser
- [ ] T023: Implement template pre-validation

## Phase 7: CLI

- [ ] T024: Implement quick mode flags
- [ ] T025: Implement flag-to-scenario compilation
- [ ] T026: Implement threshold checking and exit codes

## Phase 8: Testing

- [ ] T027: Unit tests for engine types
- [ ] T028: Unit tests for HTTP driver
- [ ] T029: Unit tests for runner
- [ ] T030: Integration tests against httptest server
- [ ] T031: Verify coverage thresholds

## Phase 9: Cleanup

- [ ] T032: Run linter and formatter
- [ ] T033: Run full test suite
- [ ] T034: Update documentation
- [ ] T035: Commit and tag

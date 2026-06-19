# Contract: Reporter Interface

**Date**: 2026-06-19
**Version**: v1.0
**Stability**: Stable
**Consumers**: `pkg/engine/Runner`
**Implementations**: `pkg/reporter.ConsoleReporter`, `pkg/reporter.JSONReporter` (FR-006), `pkg/reporter.DiscardReporter`

## Interface

```go
type Reporter interface {
    // Called once at scenario start, before setup.
    // vus is the initial number of virtual users (may increase with ramp).
    OnScenarioStart(name string, vus int)

    // Called once per phase start, after VUs are spawned.
    // vus is the target VU count for this phase.
    OnPhaseStart(name string, vus int)

    // Called when a VU goroutine begins.
    // vuID is a unique identifier for the virtual user.
    OnVUStart(vuID int)

    // Called before each Step execution within a VU.
    OnStepStart(vuID int, stepName string)

    // Called after each Step execution, with the result.
    // Thread-safe: the Runner may call this concurrently from multiple VUs.
    OnStepResult(vuID int, result *Result)

    // Called when a VU goroutine ends (either normally or via fail_fast).
    OnVUEnd(vuID int)

    // Called once per phase end, after all VUs finish.
    OnPhaseEnd(name string, results []*Result)

    // Called once at scenario end, after teardown.
    OnScenarioEnd(results []*Result)
}
```

## Semantics

### Thread Safety

All methods MUST be safe for concurrent calls from multiple VU goroutines. Implementations typically use `sync.Mutex` for output serialization.

### Ordering

Events are delivered in this order per scenario:
1. `OnScenarioStart`
2. Setup executes (not reported; driver calls only)
3. For each phase:
   a. `OnPhaseStart`
   b. `OnVUStart` (per VU spawned)
   c. `OnStepStart` → `OnStepResult` (repeated per step, per iteration)
   d. `OnVUEnd` (per VU terminated)
   e. `OnPhaseEnd`
4. Teardown executes (not reported)
5. `OnScenarioEnd`

### Null Object

`DiscardReporter` implements all methods as no-ops. This is the default when the engine is embedded as a library and the caller does not need live output.

## ConsoleReporter Output Format

```
▶ phase warmup
  ✓ health 200 12.345ms
  ✗ get-org 404 5.210ms status 404, want 200
◀ phase warmup done — 1 ok / 1 fail — avg 8.777ms

■ scenario api-smoke — 100 total — 99 ok / 1 fail
```

### Marker Legend

| Marker | Meaning |
|---|---|
| ▶ | Phase started |
| ◀ | Phase ended |
| ✓ | Step succeeded |
| ✗ | Step failed |
| ■ | Scenario summary |

## Version History

| Version | Date | Changes |
|---|---|---|
| v1.0 | 2026-06-19 | Initial contract. ConsoleReporter implements this. |

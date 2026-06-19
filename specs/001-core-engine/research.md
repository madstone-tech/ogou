# Research: Core Engine

**Date**: 2026-06-19
**Feature**: Core Engine

## Research Topics

### Topic 1: VU Concurrency Model

**Question**: How should virtual users be implemented? One goroutine per VU, or a worker pool?

**Decision**: One goroutine per VU.

**Rationale**:
- Matches the mental model: each VU is an independent actor with its own cookie jar and variables.
- Go goroutines are cheap (2KB initial stack, growable); 10,000 VUs is feasible on modern hardware.
- Simpler synchronization: each VU owns its state, no shared mutable data besides the Reporter (guarded by mutex).
- Worker pools introduce coordination complexity (stealing, fairness) without benefit for our use case.

**Alternatives considered**:
- Worker pool with channel dispatch: rejected — adds indirection without performance gain; VU state must be carried per-request anyway.
- Thread pool (fixed workers): rejected — breaks VU isolation; a worker handling VU 1 then VU 2 would need to swap contexts.

### Topic 2: Protocol-Agnostic Driver Interface

**Question**: What shape should the Driver interface take to support HTTP today and WebSocket/gRPC tomorrow?

**Decision**:
```go
type Driver interface {
    Name() string
    Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error)
    Close() error
}
```

**Rationale**:
- `Execute` takes a `Step` (method, path, headers, body) and returns a `Result`. The driver interprets these fields protocol-specifically.
- For HTTP: path is relative URL, method is HTTP verb, body is payload.
- For WebSocket (future): path is endpoint, method is message type, body is message payload.
- `Close` lets drivers clean up persistent connections (WebSocket, gRPC channels).

**Alternatives considered**:
- Per-protocol step types (`HTTPStep`, `WSStep`): rejected — adds complexity; generic Step with driver-specific interpretation is cleaner.
- Protocol negotiation in Step: rejected — over-engineering for v1; explicit driver selection is simpler.

### Topic 3: Template Engine

**Question**: Should templates use `text/template` or a custom parser?

**Decision**: `text/template` from Go stdlib.

**Rationale**:
- Already in stdlib, well-tested, secure (no arbitrary code execution in templates).
- Funcmap allows custom functions (UUID, RandomString) without parser changes.
- Familiar to Go developers; no learning curve.
- Performance is adequate for request body/path interpolation (not a hot path).

**Alternatives considered**:
- Custom template syntax: rejected — reinventing the wheel; no compelling benefit.
- `html/template`: rejected — escape behavior is wrong for API payloads (would escape JSON).

### Topic 4: Percentile Calculation

**Question**: How should percentiles be calculated?

**Decision**: Naive sort-based for v1.

**Rationale**:
- Smoke tests produce small result sets (hundreds to thousands). Sorting is O(n log n), perfectly fine.
- tdigest/HDR Histogram are O(1) insert but add dependencies and complexity.
- Deferred to FR-006 when we tackle high-throughput load testing with millions of requests.

**Alternatives considered**:
- tdigest: rejected — overkill for v1 data sizes.
- HDR Histogram: rejected — designed for continuous metrics collection, not discrete result sets.

## Resolved Clarifications

| # | Original Question | Resolution |
|---|---|---|
| 1 | VU concurrency model | One goroutine per VU |
| 2 | Driver interface shape | `Execute(ctx, baseURL, step, sc) (*Result, error)` |
| 3 | Template engine choice | `text/template` with custom funcmap |
| 4 | Percentile algorithm | Naive sort for v1; tdigest deferred |

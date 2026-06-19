# htload Specification v1.0

## Goals

Generic, open-source load and smoke testing for HTTP APIs. Supports multi-step virtual-user flows, CRUD operations, multiple auth strategies, and deployment as CLI / library / container.

## Anti-Goals

- Not a browser automation tool (no JS execution).
- Not a gRPC/WebSocket tester in v1 (driver interface reserved).
- Not specific to any single application.

---

## 1. Execution Model: Virtual User (VU)

A **scenario** is executed by N **virtual users** (VUs), each an isolated goroutine.

The lifecycle is:
1. **Global setup** runs **once** before any VU starts. Captured variables are shared across all VUs.
2. Each **phase** runs sequentially. VUs are spawned according to the phase's `rate`, loop over the phase's `steps` for the phase `duration`, then are torn down.
3. **Global teardown** runs **once** after all phases and VUs complete. Variables captured during setup are available.

```
Global Setup (once):
  login → capture token → shared vars

Phase "warmup" (5 VUs for 30s):
  VU 1: loop stepA → stepB → stepA → stepB ...
  VU 2: loop stepA → stepB → stepA → stepB ...
  ...

Phase "stress" (50 VUs for 60s):
  VU 1..50: loop stepA → stepB ...

Global Teardown (once):
  cleanup using shared token
```

State (cookies, captured variables) is scoped to the **VU** for phase execution.
Variables captured during **global setup** are copied into every VU's context at the start of each phase.

### Concurrency vs Rate

| Config | Meaning |
|---|---|
| `rate.constant: 5` | Exactly 5 VUs running concurrently |
| `rate.ramp: { from: 1, to: 10 }` | Ramps VU count linearly over phase duration |

The number of requests per second emerges from (VU count × steps per loop × iteration frequency). The user controls VU concurrency, not RPS directly.

### Multiple Phases

A scenario may declare multiple phases. They execute **sequentially** (not in parallel):

```yaml
phases:
  - name: warmup
    duration: 30s
    rate: { constant: 5 }
    steps: [...]
  - name: stress
    duration: 60s
    rate: { ramp: { from: 10, to: 50 } }
    steps: [...]
  - name: cooldown
    duration: 30s
    rate: { constant: 1 }
    steps: [...]
```

---

## 2. Scenario YAML Schema

### Top Level

```yaml
name: string              # required
base_url: string          # required
headers: map<string,string>  # global headers for all requests

cookies:                  # optional: pre-seed cookies per VU
  persistent: true      # retain cookies across steps (jar mode)

setup:                    # optional: run ONCE globally before any VU
  - step...

phases:                   # required: at least one
  - name: string
    duration: duration    # e.g., 30s, 5m
    rate:
      constant: int       # or
      ramp:
        from: int
        to: int
    steps:                # required
      - step...

teardown:                 # optional: run ONCE globally after all phases
  - step...
```

### Step

```yaml
name: string
method: string            # GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
path: string              # relative to base_url; supports templates
headers: map<string,string>  # merged with global; per-request overrides
body:                     # optional
  inline: string          # raw body
  file: string            # path to file (relative to working dir or absolute)
  template: string        # evaluated per-request with go text/template
assertions:               # optional
  - status: int
  - jsonpath:
      path: string
      equals: any           # scalar comparison
      exists: bool          # check presence
  - response_time:
      less_than: duration
captures:                 # optional: extract values into VU context
  - name: string
    source: jsonpath | header | regex
    from: string            # jsonpath expr, header name, or regex pattern
fail_fast: bool           # optional (default false). If true, stop this VU on first failure.
```

### Template Variables

All `path`, `body.inline`, `body.template`, `headers` values support `text/template`:

| Function | Example | Description |
|---|---|---|
| `{{.VU.ID}}` | `user-{{.VU.ID}}` | Unique virtual user identifier (int) |
| `{{.VU.Iteration}}` | `req-{{.VU.Iteration}}` | Loop counter for this VU |
| `{{.Vars.foo}}` | `/orgs/{{.Vars.orgID}}` | Value captured in a previous step |
| `{{.Env.KEY}}` | `{{.Env.API_BASE_URL}}` | Environment variable |
| `{{.UUID}}` | `lt-{{.UUID}}` | Random UUID v4 |
| `{{.RandomString 8}}` | `{{.RandomString 8}}` | Random alphanumeric string |
| `{{.Timestamp}}` | `{{.Timestamp}}` | Unix timestamp (seconds) |
| `{{.TimestampNano}}` | `{{.TimestampNano}}` | Unix timestamp (nanoseconds) |

### Auth Block (Optional, Top-Level)

```yaml
auth:
  strategy: none | bearer | apikey | basic | oauth-client-credentials
  # strategy-specific config follows
```

| Strategy | Fields |
|---|---|
| `none` | (default) No auth; public endpoints only. |
| `bearer` | `token: string` or `token_from_env: VAR_NAME`. Sent as `Authorization: Bearer <token>`. |
| `apikey` | `header: string` (default `X-API-Key`), `key: string` or `key_from_env: VAR_NAME`. |
| `basic` | `username: string`, `password: string` or `*_from_env`. |
| `oauth-client-credentials` | `token_url: string`, `client_id: string`, `client_secret: string`, `scope: string` (optional). Tool fetches token before each phase. |

Auth tokens are injected into every request automatically unless a step explicitly overrides the header.

**Note:** For flows requiring a login step that returns a token (e.g., `POST /login → capture JWT → use in subsequent requests`), this is handled via `setup` + `captures`, not the `auth` block. The `auth` block is for static/pre-known credentials.

---

## 3. Quick Mode (CLI)

```bash
htload <url> [flags]
```

Single endpoint, repeated for duration/count. No multi-step, no captures.

| Flag | Description | Default |
|---|---|---|
| `-c, --workers` | Number of concurrent VUs | `1` |
| `-n, --count` | Total requests (all VUs combined) | unlimited (uses `-d`) |
| `-d, --duration` | How long to run | `30s` |
| `-X, --method` | HTTP method | `GET` |
| `-H, --header` | Repeatable: `-H "Key: Value"` | none |
| `-D, --data` | Request body. Raw string or `@file` to load from file. | none |
| `-v, --verbose` | Per-request output to stderr | false |
| `--fail-if-p99` | Exit 1 if p99 > threshold | 0 (disabled) |
| `--fail-if-rate` | Exit 1 if success rate < threshold | 0 (disabled) |
| `-o, --output` | Write JSON report to file | stdout summary only |

Examples:

```bash
# GET smoke test
htload https://api.example.com/health -c 10 -n 1000

# POST load test with inline body
htload https://api.example.com/users \
  -X POST \
  -D '{"email":"test@example.com"}' \
  -H "Content-Type: application/json" \
  -c 20 -d 60s

# POST with body from file (curl convention)
htload https://api.example.com/users \
  -X POST \
  -D @payload.json \
  -H "Content-Type: application/json" \
  -c 20 -d 60s

# With threshold
htload https://api.example.com/health \
  -c 50 -d 5m \
  --fail-if-p99 500ms \
  --fail-if-rate 0.99
```

Quick mode internally compiles to a single-phase, single-step scenario.

---

## 4. Library API

```go
package htload

import (
    "context"
    "time"

    "github.com/madstone-tech/ogou/pkg/engine"
    "github.com/madstone-tech/ogou/pkg/http"
    "github.com/madstone-tech/ogou/pkg/reporter"
)

func Example() {
    scenario := engine.Scenario{
        Name:    "signup-flow",
        BaseURL: "https://api.example.com",
        Setup: []engine.Step{
            {
                Name:   "register",
                Method: "POST",
                Path:   "/api/v1/auth/signup",
                Body:   &engine.BodySource{Inline: `{"email":"test@example.com"}`},
                Captures: []engine.Capture{
                    {Name: "token", Source: "jsonpath", From: "$.token"},
                },
            },
        },
        Phases: []engine.Phase{
            {
                Name:     "warmup",
                Duration: 30 * time.Second,
                Rate:     engine.RateProfile{Constant: intPtr(5)},
                Steps: []engine.Step{
                    {
                        Name:   "list-orgs",
                        Method: "GET",
                        Path:   "/api/v1/orgs",
                        Headers: map[string]string{
                            "Authorization": "Bearer {{.Vars.token}}",
                        },
                        Assertions: []engine.Assertion{
                            {Status: intPtr(200)},
                        },
                    },
                },
            },
            {
                Name:     "stress",
                Duration: 60 * time.Second,
                Rate:     engine.RateProfile{Ramp: &engine.Ramp{From: 10, To: 50}},
                Steps: []engine.Step{
                    {
                        Name:   "create-org",
                        Method: "POST",
                        Path:   "/api/v1/orgs",
                        Headers: map[string]string{
                            "Authorization": "Bearer {{.Vars.token}}",
                        },
                        Body: &engine.BodySource{Inline: `{"name":"Org-{{.RandomString 6}}"}`},
                        Assertions: []engine.Assertion{
                            {Status: intPtr(201)},
                        },
                    },
                },
            },
        },
        Teardown: []engine.Step{
            {
                Name:   "delete-user",
                Method: "DELETE",
                Path:   "/api/v1/users/me",
                Headers: map[string]string{
                    "Authorization": "Bearer {{.Vars.token}}",
                },
            },
        },
    }

    driver := http.NewDriver(30 * time.Second)
    rep := reporter.NewConsoleReporter()
    runner := engine.NewRunner(driver, rep)

    results, err := runner.Run(context.Background(), scenario)
    // results contains every Request/Response/Assertion
}
```

The `pkg/` packages are the public API. `internal/` is implementation detail.

---

## 5. Reporter Interface

```go
type Reporter interface {
    OnScenarioStart(name string, vus int)
    OnPhaseStart(name string, vus int)
    OnVUStart(vuID int)
    OnStepStart(vuID int, stepName string)
    OnStepResult(vuID int, result *engine.Result)
    OnVUEnd(vuID int)
    OnPhaseEnd(name string, summary *engine.PhaseSummary)
    OnScenarioEnd(summary *engine.ScenarioSummary)
}
```

Built-in reporters:
- `ConsoleReporter` — TTY-friendly live output + summary table.
- `JSONReporter` — Machine-readable, suitable for `-o results.json`.
- `DiscardReporter` — Silent; for library embedding.

Planned:
- `PrometheusReporter` — Push to pushgateway.
- `S3Reporter` — Upload JSON result to S3.

---

## 6. Metrics & Exit Codes

| Metric | Description |
|---|---|
| `total_requests` | Total steps executed across all VUs |
| `successful_requests` | Steps where all assertions passed |
| `failed_requests` | Steps where any assertion or transport failed |
| `success_rate` | `successful / total` |
| `min/mean/median/p95/p99/max_latency` | Step latency distribution |
| `vus` | Virtual users configured |
| `phase_duration` | Actual wall-clock duration |
| `failures_by_step` | Count of failures per step name |

**Exit Codes:**

| Code | Meaning |
|---|---|
| `0` | All requests succeeded, all thresholds met. |
| `1` | At least one assertion failed, or a threshold (`--fail-if-*`) was breached, or runtime error. |
| `2` | Invalid scenario or CLI args. |

---

## 7. Deployment Targets

### CLI Binary
```bash
go install github.com/madstone-tech/ogou/cmd/htload@latest
htload https://api.example.com/health -c 10 -n 1000
```

### Docker
```dockerfile
FROM alpine
COPY htload /usr/local/bin/
ENTRYPOINT ["htload"]
```

```bash
docker run --rm htload https://api.example.com/health -c 10 -n 1000
```

### Library (Embedded)
Import `github.com/madstone-tech/ogou/pkg/engine` and drive programmatically (see §4).

**Out of scope for v1:**
- Lambda runtime handler
- Kubernetes CronJob manifests
- GitHub Action wrapper

These are v1.1, after the library and CLI prove the core engine.

---

## 8. Resolved Decisions

| # | Question | Decision |
|---|---|---|
| 1 | Multiple phases per scenario? | **Yes.** Phases execute sequentially. Enables warmup → stress → cooldown patterns. |
| 2 | Setup runs per-VU or once globally? | **Once globally.** Shared token/auth across all VUs. One login per scenario. |
| 3 | Capture sources for v1? | **Jsonpath + header + regex.** All three from day one. |
| 4 | Quick mode file bodies? | **Yes.** `-D @file.json` loads request body from file (curl convention). |
| 5 | Assertion failure behavior? | **Continue by default.** Optional `fail_fast: true` per step halts that VU. |

---

## 9. Directory Layout

```
cmd/htload/          CLI entry (cobra)
pkg/engine/          Public API: Scenario, Phase, Step, Runner, Reporter interface
pkg/http/            Public API: HTTP driver
pkg/reporter/        Public API: Console, JSON, Discard reporters
pkg/scenario/        Public API: YAML parser, validator
pkg/auth/            Public API: Auth provisioners (none, bearer, apikey, basic, oauth-cc)
internal/template/   Template engine: funcmap (UUID, RandomString, Env)
internal/metrics/    Latency aggregation, percentile calculation
internal/capture/    Jsonpath, header, regex extractors
internal/assertion/  Assertion engine
internal/driver/     Driver registry (HTTP default, extensible)
tests/               Integration tests against httpbin or local test server
examples/            Sample YAML scenarios
Dockerfile
SPEC.md              (this document)
```

---

*Spec version: v1.0*
*Last updated: 2026-06-19*

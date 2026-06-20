# Contract: Quick Mode

**Date**: 2026-06-20
**Version**: v1.0
**Stability**: Stable
**Consumers**: `cmd/htload/main_cobra.go`, `cmd/htload/main.go`
**Implementations**: `cmd/htload/main.go` (v1)

---

## Contract 1: Flag to Scenario Compilation

### Inputs

Parsed CLI flags and positional argument:

| Field | Type | Source |
|---|---|---|
| `URL` | `string` | Positional arg[0] |
| `Method` | `string` | `--method` / `-X` (default: GET) |
| `Workers` | `int` | `--workers` / `-c` (default: 1) |
| `Duration` | `time.Duration` | `--duration` / `-d` (default: 30s) |
| `Count` | `int` | `--count` / `-n` (default: 0) |
| `Headers` | `[]string` | `--header` / `-H` (repeatable) |
| `Data` | `string` | `--data` / `-D` (default: "") |

### Output

A single-phase `engine.Scenario`:

```go
scenario := engine.Scenario{
    Name:    "quick",
    BaseURL: url.Scheme + "://" + url.Host,
    Phases: []engine.Phase{{
        Name:     "load",
        Duration: flagDuration,
        Rate:     engine.RateProfile{Constant: &flagWorkers},
        Steps: []engine.Step{{
            Name:   "request",
            Method: flagMethod,
            Path:   url.Path + "?" + url.RawQuery,
            Headers: parsedHeaders,   // map[string]string from -H flags
            Body:    bodySource,     // nil if -D omitted
            Assertions: []engine.Assertion{
                {Status: intPtr(200)}, // default
            },
        }},
    }},
}
```

### Semantics

1. **BaseURL**: Only scheme and host. Path, query, and fragment remain in `Step.Path`.
2. **Duration precedence**: If `Count > 0`, `Duration` is still set on the phase but the runner stops after `Count` total requests. The engine honors count over duration.
3. **BodySource**:
   - `-D "raw body"` → `&engine.BodySource{Inline: "raw body"}`
   - `-D @file.json` → `&engine.BodySource{File: "file.json"}` (relative to cwd; existence validated at execution time → exit code 1 if missing)
4. **Headers**: Each `-H` flag is parsed as `"Key: Value"`. Missing colon is a validation error (exit code 2).
5. **Default assertion**: `{Status: 200}`. No other assertions in quick mode v1.

---

## Contract 2: Threshold Evaluation

### Inputs

| Field | Type | Description |
|---|---|---|
| `results` | `[]*engine.Result` | All results from the runner |
| `thresholdP99` | `time.Duration` | `--fail-if-p99` value (0 = disabled) |
| `thresholdRate` | `float64` | `--fail-if-rate` value (0 = disabled) |

### Output

| Field | Type | Description |
|---|---|---|
| `breached` | `bool` | `true` if any threshold is violated |
| `reason` | `string` | Human-readable explanation of the breach |

### Algorithm

```
ok = count of results where Result.Success == true
total = len(results)
successRate = ok / total  (if total > 0, else 0)
p99 = percentile(latencies, 99)

breached = false
reason = ""

if thresholdRate > 0 && successRate < thresholdRate:
    breached = true
    reason = "success rate X below threshold Y"

if thresholdP99 > 0 && p99 > thresholdP99:
    breached = true
    reason = "p99 X exceeds threshold Y"
```

### Semantics

1. **Percentile**: Computed from `Result.Latency` of all results, successful and failed.
2. **Success rate**: Computed as `ok / total` where `ok` is the count of `Result.Success == true`.
3. **Multiple thresholds**: Both are evaluated; if either breaches, `breached = true`. The `reason` string MUST mention all breached thresholds.
4. **Disabled thresholds**: A value of `0` means the threshold is not checked.

---

## Contract 3: Exit Code Determination

### Inputs

| Field | Type | Description |
|---|---|---|
| `argErr` | `error` | Flag parsing or validation error |
| `runErr` | `error` | Runner execution error |
| `results` | `[]*engine.Result` | All results |
| `thresholdBreached` | `bool` | From Contract 2 |

### Output

| Code | Condition |
|---|---|
| 0 | No `argErr`, no `runErr`, `thresholdBreached == false`, and at least one result exists |
| 1 | No `argErr`, but (`runErr != nil` OR `thresholdBreached == true` OR no results were produced) |
| 2 | `argErr != nil` (invalid URL, missing file for `-D @file`, malformed header, unknown flag) |

### Decision Table

| `argErr` | `runErr` | `thresholdBreached` | `len(results)` | Exit Code |
|---|---|---|---|---|
| != nil | * | * | * | 2 |
| nil | != nil | * | * | 1 |
| nil | nil | true | * | 1 |
| nil | nil | false | == 0 | 1 |
| nil | nil | false | > 0 | 0 |

### Semantics

1. **Argument errors (code 2)** take precedence over everything else. If the user passes nonsense, we tell them immediately.
2. **Runtime errors (code 1)** include: runner panics/recoveries, driver failures, threshold breaches, or zero results (which implies nothing was tested).
3. **Success (code 0)** requires at least one request was executed, all assertions passed, and no thresholds were breached.
4. The exit code is the sole machine-readable contract for CI integration. Human-readable detail goes to stderr.

## Version History

| Version | Date | Changes |
|---|---|---|
| v1.0 | 2026-06-20 | Initial contract. Covers flag→scenario compilation, threshold evaluation, and exit codes. |

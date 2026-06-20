# Quickstart: CLI Quick Mode

**Date**: 2026-06-20
**Feature**: CLI Quick Mode

## Prerequisites

- `go build ./cmd/htload` succeeds
- A running HTTP endpoint to test, or use `httptest` locally

## Basic Usage

Load-test a URL with 10 workers for 30 seconds:

```bash
htload https://example.com/api -c 10 -d 30s
```

Send exactly 100 requests:

```bash
htload https://example.com/api -n 100
```

## POST with Body and Headers

Inline JSON body:

```bash
htload https://example.com/post \
  -X POST \
  -D '{"a":1}' \
  -H "Content-Type: application/json"
```

Load body from file:

```bash
htload https://example.com/upload \
  -X POST \
  -D @payload.json \
  -H "Content-Type: application/octet-stream"
```

## Threshold-Based Exit Codes

Fail the build if p99 exceeds 100 ms:

```bash
htload https://example.com/api --fail-if-p99 100ms
```

Fail if success rate drops below 99%:

```bash
htload https://example.com/api --fail-if-rate 0.99
```

Combine both:

```bash
htload https://example.com/api \
  -c 50 -d 60s \
  --fail-if-p99 100ms \
  --fail-if-rate 0.99
```

## Exit Codes

| Code | Meaning | Typical Cause |
|---|---|---|
| 0 | Success | All requests passed assertions and thresholds |
| 1 | Failure | Assertion failed, threshold breached, connection error, or zero results |
| 2 | Invalid arguments | Bad URL, missing file, malformed header, unknown flag, negative duration |

## Full Flag Reference

| Flag | Short | Default | Description |
|---|---|---|---|
| `--workers` | `-c` | 1 | Concurrent virtual users |
| `--count` | `-n` | 0 | Total request cap (0 = duration-based) |
| `--duration` | `-d` | 30s | Test duration |
| `--method` | `-X` | GET | HTTP method |
| `--header` | `-H` | — | Repeatable header (`Key: Value`) |
| `--data` | `-D` | "" | Request body (`inline` or `@file`) |
| `--fail-if-p99` | | 0 | Exit 1 if p99 latency exceeds this |
| `--fail-if-rate` | | 0 | Exit 1 if success rate falls below this (0–1) |
| `--output` | `-o` | "" | Write JSON report to file |
| `--version` | | | Print version and exit |

## Next Steps

- Write a scenario YAML and use `htload -f scenario.yaml` (FR-003)
- Add template interpolation to request bodies
- Integrate into CI with threshold gates

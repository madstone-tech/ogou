# htload

HTTP load and smoke testing. Two layers: quick CLI flags for ad-hoc use, scenario YAML for repeatable, multi-step flows.

## Quick Start

```bash
# Build
go build ./cmd/htload

# Quick smoke test
cd examples
../htload https://httpbin.org/get -c 10 -n 100

# Scenario-based test
../htload -f smoke.yaml
```

## CLI Usage

```bash
# Quick mode (single URL)
htload https://api.example.com/health -c 10 -n 1000 -d 30s

# Scenario mode
htload -f scenario.yaml

# Fail thresholds
htload https://api.example.com -c 20 -d 60s --fail-if-p99 500ms --fail-if-rate 0.99
```

## Architecture

```
cmd/htload         CLI entry
internal/engine    Protocol-agnostic core (scenario, runner, metrics)
internal/http      HTTP driver (default)
internal/reporter  Console reporter (JSON, Prometheus planned)
internal/scenario  YAML parser + validator
```

## Roadmap

- [x] HTTP driver
- [x] Scenario YAML parser
- [x] Quick CLI mode
- [x] Console reporter
- [ ] Body file references (`body.file`)
- [ ] Variable capture (`capture_as`, `jsonpath`)
- [ ] Template interpolation (`{{.UUID}}`, `{{.RandomString}}`)
- [ ] Auth provisioner (password, API key, OAuth client credentials)
- [ ] JSON reporter + S3 output
- [ ] Prometheus push gateway reporter
- [ ] Ramp rate profile
- [ ] Plugin driver interface (WebSocket, gRPC)

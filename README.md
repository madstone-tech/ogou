# ogou (htload)

[![CI](https://github.com/madstone-tech/ogou/actions/workflows/ci.yml/badge.svg)](https://github.com/madstone-tech/ogou/actions/workflows/ci.yml)

HTTP load and smoke testing for APIs. Two-layer design: quick CLI for ad-hoc tests, scenario YAML for complex multi-step flows.

```bash
# Quick mode — one URL, no config file
htload https://api.example.com/health -c 10 -n 1000

# Scenario mode — multi-step, multi-phase, auth-aware
htload -f scenarios/smoke.yaml
```

## Features

- **Quick mode** — `htload <url> [flags]`. Zero ceremony. Supports GET/POST/PUT/DELETE/PATCH with headers and body.
- **Scenario mode** — YAML-defined multi-step virtual-user flows with setup, multiple phases, and teardown.
- **CRUD support** — Every HTTP method. Inline bodies or `-D @file.json` (curl convention).
- **Auth** — Static bearer tokens, API keys, basic auth, OAuth client credentials. Or login-via-setup with variable capture.
- **Assertions** — Status code, jsonpath, response time thresholds.
- **Captures** — Extract values from responses and reuse in subsequent steps.
- **Template interpolation** — `{{.UUID}}`, `{{.RandomString 8}}`, `{{.Env.VAR}}`, `{{.Vars.captured}}`.
- **Multi-phase execution** — Warmup → stress → cooldown, each with its own rate and duration.
- **Thresholds** — `--fail-if-p99 500ms --fail-if-rate 0.99`. CI-friendly exit codes.
- **Reports** — Console live output, JSON file output, Prometheus pushgateway (planned).

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap madstone-tech/tap
brew install ogou
```

### Go install

```bash
go install github.com/madstone-tech/ogou/cmd/htload@latest
```

### Binary download

Download the latest release for your platform from [GitHub Releases](https://github.com/madstone-tech/ogou/releases).

```bash
tar -xzf ogou_Linux_x86_64.tar.gz
sudo mv htload /usr/local/bin/
```

### Docker

```bash
docker run --rm ghcr.io/madstone-tech/ogou:latest https://api.example.com/health -c 10
```

## Quick Start

### Smoke test a single endpoint

```bash
htload https://httpbin.org/get -c 10 -n 100
```

### Load test with thresholds

```bash
htload https://api.example.com/health \
  -c 50 -d 5m \
  --fail-if-p99 500ms \
  --fail-if-rate 0.99
```

### POST with JSON body

```bash
htload https://api.example.com/users \
  -X POST \
  -D '{"email":"test@example.com"}' \
  -H "Content-Type: application/json" \
  -c 20 -d 60s
```

### POST from file

```bash
htload https://api.example.com/users \
  -X POST \
  -D @payload.json \
  -H "Content-Type: application/json" \
  -c 20 -d 60s
```

## Scenario Mode

Write a YAML file for complex, multi-step, authenticated flows:

```yaml
name: api-smoke
base_url: https://api.example.com

auth:
  strategy: bearer
  token_from_env: API_TOKEN

phases:
  - name: warmup
    duration: 10s
    rate:
      constant: 1
    steps:
      - name: health
        method: GET
        path: /health
        assertions:
          - status: 200

  - name: load
    duration: 30s
    rate:
      ramp:
        from: 5
        to: 20
    steps:
      - name: list-orgs
        method: GET
        path: /api/v1/orgs
        assertions:
          - status: 200

      - name: create-org
        method: POST
        path: /api/v1/orgs
        body:
          template: |
            {"name":"Org-{{.RandomString 6}}","slug":"org-{{.RandomString 6}}"}
        assertions:
          - status: 201
        captures:
          - name: orgID
            source: jsonpath
            from: "$.id"

      - name: get-org
        method: GET
        path: /api/v1/orgs/{{.Vars.orgID}}
        assertions:
          - status: 200
```

Run it:

```bash
htload -f scenarios/api-smoke.yaml
```

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | All successful, thresholds met. |
| `1` | Failure or threshold breached. |
| `2` | Invalid args or scenario. |

## Library Usage

Import the engine and drive programmatically:

```go
import (
    "github.com/madstone-tech/ogou/pkg/engine"
    "github.com/madstone-tech/ogou/pkg/http"
    "github.com/madstone-tech/ogou/pkg/reporter"
)

scenario := engine.Scenario{
    Name:    "signup-flow",
    BaseURL: "https://api.example.com",
    Phases: []engine.Phase{{
        Name:     "load",
        Duration: 30 * time.Second,
        Rate:     engine.RateProfile{Constant: intPtr(10)},
        Steps: []engine.Step{{
            Name:   "create-org",
            Method: "POST",
            Path:   "/api/v1/orgs",
            Body:   &engine.BodySource{Inline: `{"name":"test"}`},
        }},
    }},
}

driver := http.NewDriver(30 * time.Second)
rep := reporter.NewConsoleReporter()
runner := engine.NewRunner(driver, rep)

results, err := runner.Run(context.Background(), scenario)
```

## Development

```bash
# Clone
git clone https://github.com/madstone-tech/ogou.git
cd ogou

# Install tools
task tools

# Run CI locally
task ci

# Run tests
task test

# Build binary
task build

# Run
./htload https://httpbin.org/get -c 10 -n 100
```

See [AGENTS.md](AGENTS.md) for contributor guidance.

## Roadmap

- [x] HTTP driver
- [x] Quick CLI mode
- [x] Scenario YAML parser
- [x] Multi-phase virtual users
- [ ] JSON response assertions (jsonpath)
- [ ] Response capture (jsonpath, header, regex)
- [ ] Template interpolation (UUID, RandomString, Env, Vars)
- [ ] Auth provisioners (bearer, apikey, basic, oauth-cc)
- [ ] JSON reporter
- [ ] Prometheus pushgateway reporter
- [ ] Docker multi-arch images
- [ ] Lambda runtime handler

## License

Apache-2.0

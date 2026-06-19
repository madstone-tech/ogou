# Quickstart: Core Engine

**Date**: 2026-06-19
**Feature**: Core Engine

## Prerequisites

- Go 1.25+
- `go mod tidy` has been run

## Build

```bash
go build ./pkg/engine
go build ./pkg/http
go build ./pkg/reporter
go build ./internal/template
go build ./internal/metrics
```

## Run a Simple Test Programmatically

```go
package main

import (
    "context"
    "fmt"
    "net/http/httptest"
    "time"

    "github.com/madstone-tech/ogou/pkg/engine"
    htloadhttp "github.com/madstone-tech/ogou/pkg/http"
    "github.com/madstone-tech/ogou/pkg/reporter"
)

func main() {
    // Create a test server
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer ts.Close()

    // Define a scenario
    scenario := engine.Scenario{
        Name:    "demo",
        BaseURL: ts.URL,
        Phases: []engine.Phase{{
            Name:     "load",
            Duration: 5 * time.Second,
            Rate:     engine.RateProfile{Constant: intPtr(5)},
            Steps: []engine.Step{{
                Name:   "health",
                Method: "GET",
                Path:   "/",
                Assertions: []engine.Assertion{
                    {Status: intPtr(200)},
                },
            }},
        }},
    }

    // Run it
    driver := htloadhttp.NewDriver(10 * time.Second)
    rep := reporter.NewConsoleReporter()
    runner := engine.NewRunner(driver, rep)

    results, err := runner.Run(context.Background(), scenario)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Completed %d requests\n", len(results))
}

func intPtr(i int) *int { return &i }
```

## Expected Output

```
▶ phase load
  ✓ health 200 1.234ms
  ✓ health 200 0.987ms
  ... (many more lines)
◀ phase load done — 25 ok / 0 fail — avg 1.105ms

■ scenario demo — 25 total — 25 ok / 0 fail
```

## Next Steps

- Add template interpolation: `Body: &engine.BodySource{Template: "id={{.VU.ID}}"}`
- Add setup for auth: `Setup: []engine.Step{loginStep}`
- Add captures: `Captures: []engine.Capture{{Name: "token", Source: "jsonpath", From: "$.token"}}`
- Add fail-fast: `FailFast: true`
- Add multi-phase: warm-up, stress, cooldown

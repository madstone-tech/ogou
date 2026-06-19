# Data Model: Core Engine

**Date**: 2026-06-19
**Feature**: Core Engine

## Entity Relationship Diagram

```
Scenario
  ├── Setup: []Step          (executed once globally)
  ├── Phases: []Phase        (executed sequentially)
  │     ├── Steps: []Step     (looped per VU)
  │     └── Rate: RateProfile  (controls VU concurrency)
  └── Teardown: []Step        (executed once globally)

Phase
  └── N × VU goroutine
        └── StepContext (isolated per VU)
              ├── Vars: map[string]any
              └── Iteration: int

Step
  ├── Body: BodySource
  ├── Assertions: []Assertion
  └── Captures: []Capture

Result (produced per Step execution)
  ├── Latency: time.Duration
  ├── StatusCode: int
  ├── Success: bool
  └── Error: string
```

## Go Struct Definitions

### Scenario

```go
type Scenario struct {
    Name    string
    BaseURL string
    Headers map[string]string
    Setup   []Step
    Phases  []Phase
    Teardown []Step
}
```

### Phase

```go
type Phase struct {
    Name     string
    Duration time.Duration
    Rate     RateProfile
    Steps    []Step
}
```

### Step

```go
type Step struct {
    Name       string
    Method     string
    Path       string
    Headers    map[string]string
    Body       *BodySource
    Assertions []Assertion
    Captures   []Capture
    FailFast   bool
}
```

### BodySource

```go
type BodySource struct {
    Inline   string `yaml:"inline,omitempty"`
    File     string `yaml:"file,omitempty"`
    Template string `yaml:"template,omitempty"`
}

func (b *BodySource) Body() ([]byte, error) { ... }
```

### RateProfile and Ramp

```go
type RateProfile struct {
    Constant *int  `yaml:"constant,omitempty"`
    Ramp     *Ramp `yaml:"ramp,omitempty"`
}

type Ramp struct {
    From int `yaml:"from"`
    To   int `yaml:"to"`
}
```

### Assertion

```go
type Assertion struct {
    Status       *int
    JSONPath     *JSONPathAssertion
    ResponseTime *ResponseTimeAssertion
}

type JSONPathAssertion struct {
    Path      string
    Equals    interface{}
    Exists    bool
    CaptureAs string
}

type ResponseTimeAssertion struct {
    LessThan time.Duration
}
```

### Capture

```go
type Capture struct {
    Name   string
    Source string // "jsonpath" | "header" | "regex"
    From   string
}
```

### Result

```go
type Result struct {
    StepName     string
    StartTime    time.Time
    EndTime      time.Time
    Latency      time.Duration
    StatusCode   int
    Success      bool
    Error        string
    CapturedVars map[string]any
}
```

### StepContext

```go
type StepContext struct {
    Vars      map[string]any
    Iteration int
}

func NewStepContext() *StepContext {
    return &StepContext{Vars: make(map[string]any)}
}
```

## Validation Rules

| Entity | Field | Rule |
|---|---|---|
| Scenario | Name | Required, non-empty |
| Scenario | BaseURL | Required, valid URL |
| Scenario | Phases | Required, at least one |
| Phase | Name | Required |
| Phase | Duration | Must be > 0 |
| Phase | Rate | Exactly one of Constant or Ramp must be set |
| Phase | Steps | Required, at least one |
| Step | Name | Required |
| Step | Path | Required |
| Step | Method | Defaults to "GET" if empty |
| BodySource | — | Exactly one of Inline, File, Template may be set |
| RateProfile | — | Exactly one of Constant or Ramp must be set |

## State Transitions

### Runner Lifecycle

```
[Idle] → OnScenarioStart
       → Run Setup (once)
       → For each Phase:
           → OnPhaseStart
           → Spawn VUs
           → VUs loop Steps
           → Wait for VUs
           → OnPhaseEnd
       → Run Teardown (once)
       → OnScenarioEnd → [Complete]
```

### VU Lifecycle

```
Spawned → Receive shared vars from setup
       → Loop:
           → Increment iteration
           → For each Step:
               → Interpolate templates
               → Driver.Execute
               → Run assertions
               → Apply captures
               → Reporter.OnStepResult
               → If FailFast && !Success: exit loop
           → If phase duration expired: exit loop
       → Done
```

// Package engine defines the protocol-agnostic core of htload.
// It is intentionally independent of transport details.
package engine

import "time"

// Scenario is the top-level unit of execution.
type Scenario struct {
	Name     string            `yaml:"name"`
	BaseURL  string            `yaml:"base_url"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Setup    []Step            `yaml:"setup,omitempty"`
	Phases   []Phase           `yaml:"phases"`
	Teardown []Step            `yaml:"teardown,omitempty"`
}

// Phase is a stage with a specific rate profile and duration.
type Phase struct {
	Name     string        `yaml:"name"`
	Duration time.Duration `yaml:"duration"`
	Rate     RateProfile   `yaml:"rate"`
	Steps    []Step        `yaml:"steps"`
}

// RateProfile describes how many virtual users/requests to generate.
// Only Constant is implemented today; Ramp and Spike are reserved.
type RateProfile struct {
	Constant *int  `yaml:"constant,omitempty"`
	Ramp     *Ramp `yaml:"ramp,omitempty"`
}

// Ramp ramps from From to To concurrent users over the phase duration.
type Ramp struct {
	From int `yaml:"from"`
	To   int `yaml:"to"`
}

// Step is a single request to issue.
type Step struct {
	Name       string            `yaml:"name"`
	Method     string            `yaml:"method"`
	Path       string            `yaml:"path"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	Body       *BodySource       `yaml:"body,omitempty"`
	Assertions []Assertion       `yaml:"assertions,omitempty"`
	Captures   []Capture         `yaml:"captures,omitempty"`
	FailFast   bool              `yaml:"fail_fast,omitempty"`
}

// BodySource carries a request body via inline string, file reference, or template.
type BodySource struct {
	Inline   string `yaml:"inline,omitempty"`
	File     string `yaml:"file,omitempty"`
	Template string `yaml:"template,omitempty"`
}

// Body returns the effective body bytes, reading from file if needed.
func (b *BodySource) Body() ([]byte, error) {
	if b.File != "" {
		return nil, nil // TODO: read file
	}
	return []byte(b.Inline), nil
}

// Assertion checks something about the response.
type Assertion struct {
	Status       *int                   `yaml:"status,omitempty"`
	JSONPath     *JSONPathAssertion     `yaml:"jsonpath,omitempty"`
	ResponseTime *ResponseTimeAssertion `yaml:"response_time,omitempty"`
}

// JSONPathAssertion asserts on a JSON path expression.
type JSONPathAssertion struct {
	Path      string      `yaml:"path"`
	Equals    interface{} `yaml:"equals,omitempty"`
	Exists    bool        `yaml:"exists,omitempty"`
	CaptureAs string      `yaml:"capture_as,omitempty"`
}

// ResponseTimeAssertion asserts on response latency.
type ResponseTimeAssertion struct {
	LessThan time.Duration `yaml:"less_than"`
}

// Capture extracts a value from the response into the step context.
type Capture struct {
	Name string `yaml:"name"`
	// TODO: JSONPath, Header, Regex extractors
}

// StepContext carries mutable state for a single virtual user's journey.
type StepContext struct {
	VUID      int
	Vars      map[string]any
	Iteration int
}

// NewStepContext creates a fresh context.
func NewStepContext() *StepContext {
	return &StepContext{Vars: make(map[string]any)}
}

// Result is the outcome of executing one Step.
type Result struct {
	StepName     string
	StartTime    time.Time
	EndTime      time.Time
	Latency      time.Duration
	StatusCode   int    // transport-agnostic: HTTP 200, WS 101, gRPC 0
	Success      bool   // passed all assertions
	Error        string // non-empty if transport or assertion failed
	CapturedVars map[string]any
}

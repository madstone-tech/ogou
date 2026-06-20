package engine

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockDriver is a test double for Driver.
type mockDriver struct {
	name      string
	mu        sync.Mutex
	results   map[int][]*Result // keyed by VUID
	callCount int
}

func (m *mockDriver) Name() string { return m.name }

func (m *mockDriver) Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	res := &Result{
		StepName:   step.Name,
		Success:    true,
		StatusCode: 200,
		Latency:    1 * time.Millisecond,
	}
	m.results[sc.VUID] = append(m.results[sc.VUID], res)
	return res, nil
}

func (m *mockDriver) Close() error { return nil }

// mockReporter is a test double for Reporter.
type mockReporter struct {
	mu             sync.Mutex
	scenarioStarts int
	phaseStarts    int
	phaseEnds      int
	scenarioEnds   int
	stepResults    int
}

func (m *mockReporter) OnScenarioStart(name string, vus int) {
	m.mu.Lock()
	m.scenarioStarts++
	m.mu.Unlock()
}
func (m *mockReporter) OnPhaseStart(name string, vus int) {
	m.mu.Lock()
	m.phaseStarts++
	m.mu.Unlock()
}
func (m *mockReporter) OnVUStart(vuID int)                    {}
func (m *mockReporter) OnStepStart(vuID int, stepName string) {}
func (m *mockReporter) OnStepResult(vuID int, result *Result) {
	m.mu.Lock()
	m.stepResults++
	m.mu.Unlock()
}
func (m *mockReporter) OnVUEnd(vuID int) {}
func (m *mockReporter) OnPhaseEnd(name string, results []*Result) {
	m.mu.Lock()
	m.phaseEnds++
	m.mu.Unlock()
}
func (m *mockReporter) OnScenarioEnd(results []*Result) { m.mu.Lock(); m.scenarioEnds++; m.mu.Unlock() }
func TestRunnerBasic(t *testing.T) {
	driver := &mockDriver{name: "mock", results: make(map[int][]*Result)}
	rep := &mockReporter{}
	runner := NewRunner(driver, rep)

	scenario := &Scenario{
		Name:    "test",
		BaseURL: "http://example.com",
		Phases: []Phase{{
			Name:     "p1",
			Duration: 50 * time.Millisecond,
			Rate:     RateProfile{Constant: intPtr(2)},
			Steps: []Step{{
				Name:   "s1",
				Method: "GET",
				Path:   "/",
			}},
		}},
	}

	results, err := runner.Run(context.Background(), scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if rep.scenarioStarts != 1 {
		t.Fatalf("expected 1 scenario start, got %d", rep.scenarioStarts)
	}
	if rep.phaseStarts != 1 {
		t.Fatalf("expected 1 phase start, got %d", rep.phaseStarts)
	}
	if rep.phaseEnds != 1 {
		t.Fatalf("expected 1 phase end, got %d", rep.phaseEnds)
	}
	if rep.scenarioEnds != 1 {
		t.Fatalf("expected 1 scenario end, got %d", rep.scenarioEnds)
	}
}

// failFastDriver fails every request.
type failFastDriver struct{}

func (f *failFastDriver) Name() string { return "failfast" }
func (f *failFastDriver) Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error) {
	return &Result{StepName: step.Name, Success: false, StatusCode: 500, Latency: 1 * time.Millisecond}, nil
}
func (f *failFastDriver) Close() error { return nil }

func TestRunnerFailFast(t *testing.T) {
	driver := &failFastDriver{}
	rep := &mockReporter{}
	runner := NewRunner(driver, rep)

	scenario := &Scenario{
		Name:    "test",
		BaseURL: "http://example.com",
		Phases: []Phase{{
			Name:     "p1",
			Duration: 5 * time.Second,
			Rate:     RateProfile{Constant: intPtr(1)},
			Steps: []Step{{
				Name:     "fail",
				Method:   "GET",
				Path:     "/",
				FailFast: true,
			}},
		}},
	}

	results, err := runner.Run(context.Background(), scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (fail_fast), got %d", len(results))
	}
}

func TestRunnerSetupTeardown(t *testing.T) {
	driver := &mockDriver{name: "mock", results: make(map[int][]*Result)}
	rep := &mockReporter{}
	runner := NewRunner(driver, rep)

	scenario := &Scenario{
		Name:    "test",
		BaseURL: "http://example.com",
		Setup: []Step{{
			Name:   "setup",
			Method: "GET",
			Path:   "/setup",
		}},
		Phases: []Phase{{
			Name:     "p1",
			Duration: 50 * time.Millisecond,
			Rate:     RateProfile{Constant: intPtr(1)},
			Steps:    []Step{{Name: "s1", Method: "GET", Path: "/"}},
		}},
		Teardown: []Step{{
			Name:   "teardown",
			Method: "GET",
			Path:   "/teardown",
		}},
	}

	_, err := runner.Run(context.Background(), scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver.callCount < 3 {
		t.Fatalf("expected at least 3 calls (setup + phase + teardown), got %d", driver.callCount)
	}
}

func TestRunnerPanicRecovery(t *testing.T) {
	panicDriver := &panicDriver{}
	rep := &mockReporter{}
	runner := NewRunner(panicDriver, rep)

	scenario := &Scenario{
		Name:    "test",
		BaseURL: "http://example.com",
		Phases: []Phase{{
			Name:     "p1",
			Duration: 50 * time.Millisecond,
			Rate:     RateProfile{Constant: intPtr(1)},
			Steps:    []Step{{Name: "s1", Method: "GET", Path: "/"}},
		}},
	}

	results, err := runner.Run(context.Background(), scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result (panic recovery)")
	}
	foundPanic := false
	for _, r := range results {
		if r.Error != "" && r.Error[:5] == "panic" {
			foundPanic = true
			break
		}
	}
	if !foundPanic {
		t.Fatal("expected panic result")
	}
}

// panicDriver always panics.
type panicDriver struct{}

func (p *panicDriver) Name() string { return "panic" }
func (p *panicDriver) Execute(ctx context.Context, baseURL string, step Step, sc *StepContext) (*Result, error) {
	panic("intentional panic")
}
func (p *panicDriver) Close() error { return nil }

func TestNewRunner(t *testing.T) {
	d := &mockDriver{}
	r := &mockReporter{}
	runner := NewRunner(d, r)
	if runner == nil {
		t.Fatal("expected non-nil runner")
	}
}

func TestNewStepContext(t *testing.T) {
	sc := NewStepContext()
	if sc == nil {
		t.Fatal("expected non-nil context")
	}
	if sc.Vars == nil {
		t.Fatal("expected non-nil vars map")
	}
	if len(sc.Vars) != 0 {
		t.Fatal("expected empty vars")
	}
}

func TestBodySourceBody(t *testing.T) {
	bs := &BodySource{Inline: "hello"}
	b, err := bs.Body()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("expected 'hello', got %s", string(b))
	}
}

func intPtr(i int) *int {
	return &i
}

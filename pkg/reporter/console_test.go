package reporter

import (
	"testing"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
)

func TestConsoleReporterNoPanic(t *testing.T) {
	rep := NewConsoleReporter()
	rep.OnScenarioStart("test", 1)
	rep.OnPhaseStart("warmup", 1)
	rep.OnVUStart(0)
	rep.OnStepStart(0, "health")
	rep.OnStepResult(0, &engine.Result{
		StepName:   "health",
		StatusCode: 200,
		Latency:    10 * time.Millisecond,
		Success:    true,
	})
	rep.OnStepResult(0, &engine.Result{
		StepName:   "fail",
		StatusCode: 500,
		Latency:    5 * time.Millisecond,
		Success:    false,
		Error:      "status 500, want 200",
	})
	rep.OnVUEnd(0)
	rep.OnPhaseEnd("warmup", []*engine.Result{
		{Success: true},
		{Success: false},
	})
	rep.OnScenarioEnd([]*engine.Result{
		{Success: true},
		{Success: false},
	})
	// If we get here without panic, the reporter is functional.
}

func TestConsoleReporterSummary(t *testing.T) {
	rep := NewConsoleReporter()
	results := []*engine.Result{
		{Success: true, Latency: 10 * time.Millisecond},
		{Success: true, Latency: 20 * time.Millisecond},
		{Success: false, Latency: 30 * time.Millisecond},
	}
	rep.OnPhaseEnd("load", results)
	rep.OnScenarioEnd(results)
	// No panic = pass
}

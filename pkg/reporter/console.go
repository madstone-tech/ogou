package reporter

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
)

// ConsoleReporter prints live progress to stderr.
type ConsoleReporter struct {
	mu  sync.Mutex
	log *slog.Logger
}

// NewConsoleReporter creates a new console reporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		log: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
}

func (c *ConsoleReporter) OnScenarioStart(name string, vus int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintf(os.Stderr, "▶ scenario %s (%d VUs)\n", name, vus)
	c.log.Info("scenario_start", "name", name, "vus", vus)
}

func (c *ConsoleReporter) OnPhaseStart(name string, vus int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Fprintf(os.Stderr, "▶ phase %s (%d VUs)\n", name, vus)
	c.log.Info("phase_start", "name", name, "vus", vus)
}

func (c *ConsoleReporter) OnVUStart(vuID int) {
	// Intentionally quiet; log at debug level only
}

func (c *ConsoleReporter) OnStepStart(vuID int, stepName string) {
	// Intentionally quiet; avoids flooding output
}

func (c *ConsoleReporter) OnStepResult(vuID int, r *engine.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	marker := "✓"
	if !r.Success || r.Error != "" {
		marker = "✗"
	}
	fmt.Fprintf(os.Stderr, "  %s %s %d %s %s\n",
		marker, r.StepName, r.StatusCode, r.Latency, r.Error)
	c.log.Info("step_result",
		"vu_id", vuID,
		"step", r.StepName,
		"status", r.StatusCode,
		"latency_ms", r.Latency.Milliseconds(),
		"success", r.Success,
		"error", r.Error,
	)
}

func (c *ConsoleReporter) OnVUEnd(vuID int) {
	// Intentionally quiet
}

func (c *ConsoleReporter) OnPhaseEnd(name string, results []*engine.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var ok, fail int
	var totalDur time.Duration
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			fail++
		}
		totalDur += r.Latency
	}
	avg := time.Duration(0)
	if len(results) > 0 {
		avg = totalDur / time.Duration(len(results))
	}
	fmt.Fprintf(os.Stderr, "◀ phase %s done — %d ok / %d fail — avg %s\n", name, ok, fail, avg)
	c.log.Info("phase_end", "name", name, "ok", ok, "fail", fail, "avg_ms", avg.Milliseconds())
}

func (c *ConsoleReporter) OnScenarioEnd(results []*engine.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var ok, fail int
	var totalDur time.Duration
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			fail++
		}
		totalDur += r.Latency
	}
	avg := time.Duration(0)
	if len(results) > 0 {
		avg = totalDur / time.Duration(len(results))
	}
	fmt.Fprintf(os.Stderr, "■ scenario done — %d total — %d ok / %d fail — avg %s\n",
		len(results), ok, fail, avg)
	c.log.Info("scenario_end", "total", len(results), "ok", ok, "fail", fail, "avg_ms", avg.Milliseconds())
}

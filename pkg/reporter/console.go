package reporter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
)

// ConsoleReporter prints live progress to stderr.
type ConsoleReporter struct {
	mu sync.Mutex
}

// NewConsoleReporter creates a new console reporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (c *ConsoleReporter) OnPhaseStart(name string) {
	fmt.Fprintf(os.Stderr, "▶ phase %s\n", name)
}

func (c *ConsoleReporter) OnResult(r *engine.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	marker := "✓"
	if !r.Success || r.Error != "" {
		marker = "✗"
	}
	fmt.Fprintf(os.Stderr, "  %s %s %d %s %s\n",
		marker, r.StepName, r.StatusCode, r.Latency, r.Error)
}

func (c *ConsoleReporter) OnPhaseEnd(name string, results []*engine.Result) {
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
}

func (c *ConsoleReporter) OnScenarioEnd(name string, results []*engine.Result) {
	var ok, fail int
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			fail++
		}
	}
	fmt.Fprintf(os.Stderr, "■ scenario %s — %d total — %d ok / %d fail\n", name, len(results), ok, fail)
}

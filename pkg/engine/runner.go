package engine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Runner executes a Scenario against a Driver.
type Runner struct {
	driver   Driver
	reporter Reporter
}

// NewRunner creates a new runner.
func NewRunner(driver Driver, reporter Reporter) *Runner {
	return &Runner{driver: driver, reporter: reporter}
}

// Run executes the scenario to completion.
func (r *Runner) Run(ctx context.Context, s *Scenario) ([]*Result, error) {
	var allResults []*Result

	for _, phase := range s.Phases {
		r.reporter.OnPhaseStart(phase.Name)

		results, err := r.runPhase(ctx, s.BaseURL, phase)
		if err != nil {
			return allResults, fmt.Errorf("phase %q: %w", phase.Name, err)
		}
		allResults = append(allResults, results...)
		r.reporter.OnPhaseEnd(phase.Name, results)
	}

	r.reporter.OnScenarioEnd(s.Name, allResults)
	return allResults, nil
}

func (r *Runner) runPhase(ctx context.Context, baseURL string, phase Phase) ([]*Result, error) {
	// Determine concurrency: use constant if set, else ramp midpoint.
	concurrency := 1
	if phase.Rate.Constant != nil {
		concurrency = *phase.Rate.Constant
	} else if phase.Rate.Ramp != nil {
		concurrency = phase.Rate.Ramp.From
	}

	deadline, cancel := context.WithDeadline(ctx, time.Now().Add(phase.Duration))
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan *Result, 1000)

	// Spawn workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sc := NewStepContext()
			iteration := 0
			for {
				select {
				case <-deadline.Done():
					return
				default:
				}
				for _, step := range phase.Steps {
					sc.Iteration = iteration
					res, err := r.driver.Execute(deadline, baseURL, step, sc)
					if err != nil {
						res = &Result{StepName: step.Name, Error: err.Error(), Success: false}
					}
					select {
					case results <- res:
					case <-deadline.Done():
						return
					}
				}
				iteration++
			}
		}(i)
	}

	// Collector
	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []*Result
	for res := range results {
		r.reporter.OnResult(res)
		collected = append(collected, res)
	}
	return collected, nil
}

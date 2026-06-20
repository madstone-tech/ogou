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
	sharedVars := make(map[string]any)

	// Determine total VU count for OnScenarioStart
	totalVUs := 0
	for _, phase := range s.Phases {
		if phase.Rate.Constant != nil {
			totalVUs += *phase.Rate.Constant
		} else if phase.Rate.Ramp != nil {
			totalVUs += phase.Rate.Ramp.To
		}
	}
	r.reporter.OnScenarioStart(s.Name, totalVUs)

	// Global setup
	if len(s.Setup) > 0 {
		setupCtx := NewStepContext()
		for _, step := range s.Setup {
			res, err := r.driver.Execute(ctx, s.BaseURL, step, setupCtx)
			if err != nil {
				res = &Result{StepName: step.Name, Error: err.Error(), Success: false}
			}
			for k, v := range res.CapturedVars {
				sharedVars[k] = v
			}
			if !res.Success {
				return nil, fmt.Errorf("setup step %q failed: %s", step.Name, res.Error)
			}
		}
	}

	for _, phase := range s.Phases {
		results, err := r.runPhase(ctx, s.BaseURL, phase, sharedVars)
		if err != nil {
			return allResults, fmt.Errorf("phase %q: %w", phase.Name, err)
		}
		allResults = append(allResults, results...)
	}

	// Global teardown
	if len(s.Teardown) > 0 {
		teardownCtx := NewStepContext()
		teardownCtx.Vars = sharedVars
		for _, step := range s.Teardown {
			res, err := r.driver.Execute(ctx, s.BaseURL, step, teardownCtx)
			if err != nil {
				res = &Result{StepName: step.Name, Error: err.Error(), Success: false}
			}
			if !res.Success {
				return allResults, fmt.Errorf("teardown step %q failed: %s", step.Name, res.Error)
			}
		}
	}

	r.reporter.OnScenarioEnd(allResults)
	return allResults, nil
}

func (r *Runner) runPhase(ctx context.Context, baseURL string, phase Phase, sharedVars map[string]any) ([]*Result, error) {
	concurrency := 1
	if phase.Rate.Constant != nil {
		concurrency = *phase.Rate.Constant
	} else if phase.Rate.Ramp != nil {
		concurrency = phase.Rate.Ramp.From
	}

	r.reporter.OnPhaseStart(phase.Name, concurrency)

	deadline, cancel := context.WithDeadline(ctx, time.Now().Add(phase.Duration))
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan *Result, 1000)

	// Spawn workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(vuID int) {
			defer func() {
				if rec := recover(); rec != nil {
					res := &Result{
						StepName: "panic",
						Success:  false,
						Error:    fmt.Sprintf("panic: %v", rec),
					}
					select {
					case results <- res:
					case <-deadline.Done():
					}
				}
				wg.Done()
				r.reporter.OnVUEnd(vuID)
			}()

			sc := NewStepContext()
			sc.VUID = vuID
			// Prime with setup-captured variables
			for k, v := range sharedVars {
				sc.Vars[k] = v
			}

			r.reporter.OnVUStart(vuID)

			iteration := 0
			for {
				select {
				case <-deadline.Done():
					return
				default:
				}
				for _, step := range phase.Steps {
					sc.Iteration = iteration
					r.reporter.OnStepStart(vuID, step.Name)

					res, err := r.driver.Execute(deadline, baseURL, step, sc)
					if err != nil {
						res = &Result{StepName: step.Name, Error: err.Error(), Success: false}
					}
					if res == nil {
						res = &Result{StepName: step.Name, Success: false, Error: "nil result"}
					}
					for k, v := range res.CapturedVars {
						sc.Vars[k] = v
					}

					select {
					case results <- res:
					case <-deadline.Done():
						return
					}
					r.reporter.OnStepResult(vuID, res)

					if step.FailFast && !res.Success {
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
		collected = append(collected, res)
	}

	r.reporter.OnPhaseEnd(phase.Name, collected)
	return collected, nil
}

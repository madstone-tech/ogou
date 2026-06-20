package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
	htloadhttp "github.com/madstone-tech/ogou/pkg/http"
	"github.com/madstone-tech/ogou/pkg/reporter"
)

func TestRunnerEndToEnd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	scenario := engine.Scenario{
		Name:    "e2e-test",
		BaseURL: ts.URL,
		Phases: []engine.Phase{{
			Name:     "load",
			Duration: 100 * time.Millisecond,
			Rate:     engine.RateProfile{Constant: intPtr(2)},
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

	driver := htloadhttp.NewDriver(10 * time.Second)
	rep := reporter.NewConsoleReporter()
	runner := engine.NewRunner(driver, rep)

	results, err := runner.Run(context.Background(), &scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	var okCount int
	for _, r := range results {
		if r.Success {
			okCount++
		} else {
			t.Logf("failure: %s", r.Error)
		}
	}
	okRate := float64(okCount) / float64(len(results))
	if okRate < 0.80 {
		t.Fatalf("expected >=80%% success, got %.1f%% (%d/%d)", okRate*100, okCount, len(results))
	}
}

func TestRunnerFailFast(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	scenario := engine.Scenario{
		Name:    "failfast-test",
		BaseURL: ts.URL,
		Phases: []engine.Phase{{
			Name:     "load",
			Duration: 5 * time.Second,
			Rate:     engine.RateProfile{Constant: intPtr(1)},
			Steps: []engine.Step{{
				Name:     "fail",
				Method:   "GET",
				Path:     "/",
				FailFast: true,
				Assertions: []engine.Assertion{
					{Status: intPtr(200)},
				},
			}},
		}},
	}

	driver := htloadhttp.NewDriver(10 * time.Second)
	rep := reporter.NewConsoleReporter()
	runner := engine.NewRunner(driver, rep)

	results, err := runner.Run(context.Background(), &scenario)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result (fail_fast), got %d", len(results))
	}
	if results[0].Success {
		t.Fatal("expected failure")
	}
}

func intPtr(i int) *int {
	return &i
}

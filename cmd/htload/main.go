package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
	"github.com/madstone-tech/ogou/pkg/http"
	"github.com/madstone-tech/ogou/pkg/reporter"
	"github.com/madstone-tech/ogou/pkg/scenario"
	"github.com/spf13/cobra"
)

var (
	flagConfig   string
	flagDuration time.Duration
	flagCount    int
	flagWorkers  int
	flagVerbose  bool
	flagFailP99  time.Duration
	flagFailRate float64
	flagOutput   string
)

var rootCmd = &cobra.Command{
	Use:   "htload [URL]",
	Short: "HTTP load and smoke testing",
	Long: `htload is a two-layer load tester:

  Quick mode:    htload https://api.example.com/health -c 100
  Scenario mode: htload -f scenario.yaml
`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&flagConfig, "file", "f", "", "Path to scenario YAML")
	rootCmd.Flags().DurationVarP(&flagDuration, "duration", "d", 30*time.Second, "Test duration")
	rootCmd.Flags().IntVarP(&flagCount, "count", "n", 0, "Total requests to send (default: unlimited for duration)")
	rootCmd.Flags().IntVarP(&flagWorkers, "workers", "c", 1, "Concurrent workers")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().DurationVar(&flagFailP99, "fail-if-p99", 0, "Exit 1 if p99 latency exceeds this")
	rootCmd.Flags().Float64Var(&flagFailRate, "fail-if-rate", 0, "Exit 1 if success rate falls below this (0-1)")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write JSON report to file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if flagConfig != "" {
		return runScenario(flagConfig)
	}
	if len(args) == 0 {
		return fmt.Errorf("provide a URL or use --file")
	}
	return runQuick(args[0])
}

func runQuick(url string) error {
	s := &engine.Scenario{
		Name:    "quick",
		BaseURL: url,
		Phases: []engine.Phase{{
			Name:     "load",
			Duration: flagDuration,
			Rate:     engine.RateProfile{Constant: &flagWorkers},
			Steps: []engine.Step{{
				Name:   "GET",
				Method: "GET",
				Path:   "/",
			}},
		}},
	}
	return execute(s)
}

func runScenario(path string) error {
	s, err := scenario.Load(path)
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}
	return execute(s)
}

func execute(s *engine.Scenario) error {
	driver := http.NewDriver(30 * time.Second)
	rep := reporter.NewConsoleReporter()
	runner := engine.NewRunner(driver, rep)

	results, err := runner.Run(context.Background(), s)
	if err != nil {
		return err
	}

	// Compute summary
	var ok, failed int
	var latencies []time.Duration
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			failed++
		}
		latencies = append(latencies, r.Latency)
	}

	total := len(results)
	successRate := 0.0
	if total > 0 {
		successRate = float64(ok) / float64(total)
	}
	p99 := percentile(latencies, 99)

	fmt.Printf("\nSummary: %d requests, %d ok, %d failed, %.2f%% success, p99 %s\n",
		total, ok, failed, successRate*100, p99)

	// Threshold checks
	if flagFailRate > 0 && successRate < flagFailRate {
		return fmt.Errorf("success rate %.2f below threshold %.2f", successRate, flagFailRate)
	}
	if flagFailP99 > 0 && p99 > flagFailP99 {
		return fmt.Errorf("p99 %s exceeds threshold %s", p99, flagFailP99)
	}

	return nil
}

func percentile(latencies []time.Duration, p int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	// naive; replace with sort + index for production
	total := time.Duration(0)
	for _, l := range latencies {
		total += l
	}
	return total / time.Duration(len(latencies)) // placeholder
}

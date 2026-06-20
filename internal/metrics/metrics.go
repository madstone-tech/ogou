// Package metrics provides latency aggregation utilities.
package metrics

import (
	"sort"
	"time"
)

// CalculateMean returns the average of a slice of durations.
func CalculateMean(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

// CalculateMin returns the minimum duration.
func CalculateMin(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}
	return min
}

// CalculateMax returns the maximum duration.
func CalculateMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

// CalculateMedian returns the median duration.
func CalculateMedian(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// CalculatePercentile returns the p-th percentile (0-100).
func CalculatePercentile(durations []time.Duration, p int) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	if p <= 0 {
		return CalculateMin(durations)
	}
	if p >= 100 {
		return CalculateMax(durations)
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(idx)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}
	fraction := idx - float64(lower)
	return sorted[lower] + time.Duration(fraction*float64(sorted[upper]-sorted[lower]))
}

package metrics

import (
	"testing"
	"time"
)

func TestCalculateMean(t *testing.T) {
	got := CalculateMean([]time.Duration{1 * time.Millisecond, 3 * time.Millisecond})
	want := 2 * time.Millisecond
	if got != want {
		t.Fatalf("mean: want %v, got %v", want, got)
	}
}

func TestCalculateMeanEmpty(t *testing.T) {
	got := CalculateMean([]time.Duration{})
	if got != 0 {
		t.Fatalf("mean of empty: want 0, got %v", got)
	}
}

func TestCalculateMinMax(t *testing.T) {
	min := CalculateMin([]time.Duration{3 * time.Millisecond, 1 * time.Millisecond, 2 * time.Millisecond})
	max := CalculateMax([]time.Duration{3 * time.Millisecond, 1 * time.Millisecond, 2 * time.Millisecond})
	if min != 1*time.Millisecond {
		t.Fatalf("min: want 1ms, got %v", min)
	}
	if max != 3*time.Millisecond {
		t.Fatalf("max: want 3ms, got %v", max)
	}
}

func TestCalculateMedian(t *testing.T) {
	odd := CalculateMedian([]time.Duration{1 * time.Millisecond, 3 * time.Millisecond, 2 * time.Millisecond})
	if odd != 2*time.Millisecond {
		t.Fatalf("median odd: want 2ms, got %v", odd)
	}
	even := CalculateMedian([]time.Duration{1 * time.Millisecond, 4 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond})
	want := (2*time.Millisecond + 3*time.Millisecond) / 2
	if even != want {
		t.Fatalf("median even: want %v, got %v", want, even)
	}
}

func TestCalculatePercentile(t *testing.T) {
	d := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
	}
	if p := CalculatePercentile(d, 0); p != 1*time.Millisecond {
		t.Fatalf("p0: want 1ms, got %v", p)
	}
	if p := CalculatePercentile(d, 50); p != 3*time.Millisecond {
		t.Fatalf("p50: want 3ms, got %v", p)
	}
	if p := CalculatePercentile(d, 100); p != 5*time.Millisecond {
		t.Fatalf("p100: want 5ms, got %v", p)
	}
	if p := CalculatePercentile(d, 25); p < 1*time.Millisecond || p > 3*time.Millisecond {
		t.Fatalf("p25: unexpected value %v", p)
	}
}

func TestCalculatePercentileEmpty(t *testing.T) {
	p := CalculatePercentile([]time.Duration{}, 50)
	if p != 0 {
		t.Fatalf("percentile of empty: want 0, got %v", p)
	}
}

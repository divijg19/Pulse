package metrics

import (
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

func TestComputeSummary(t *testing.T) {
	results := []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
		{Status: 204, Latency: 200 * time.Millisecond},
		{Status: 500, Latency: 300 * time.Millisecond},
	}

	summary := Compute(results, 1500*time.Millisecond)
	if summary.Total != 3 {
		t.Fatalf("Total = %d", summary.Total)
	}
	if summary.Successes != 2 {
		t.Fatalf("Successes = %d", summary.Successes)
	}
	if summary.SuccessRate != 67 {
		t.Fatalf("SuccessRate = %d", summary.SuccessRate)
	}
	if summary.Average != 200*time.Millisecond {
		t.Fatalf("Average = %s", summary.Average)
	}
	if summary.MinLatency != 100*time.Millisecond {
		t.Fatalf("MinLatency = %s", summary.MinLatency)
	}
	if summary.MaxLatency != 300*time.Millisecond {
		t.Fatalf("MaxLatency = %s", summary.MaxLatency)
	}
	if summary.RequestsPerS != 2 {
		t.Fatalf("RequestsPerS = %f", summary.RequestsPerS)
	}
	if summary.P50 != 200*time.Millisecond {
		t.Fatalf("P50 = %s", summary.P50)
	}
	if summary.P90 != 300*time.Millisecond {
		t.Fatalf("P90 = %s", summary.P90)
	}
	if summary.P99 != 300*time.Millisecond {
		t.Fatalf("P99 = %s", summary.P99)
	}
}

func TestComputeEmptySummary(t *testing.T) {
	summary := Compute(nil, time.Second)
	if summary.Total != 0 || summary.SuccessRate != 0 || summary.RequestsPerS != 0 {
		t.Fatalf("unexpected empty summary: %+v", summary)
	}
}

func TestComputeEdgeCases(t *testing.T) {
	t.Run("single result", func(t *testing.T) {
		results := []model.Result{
			{Status: 200, Latency: 123 * time.Millisecond},
		}
		s := Compute(results, time.Second)
		if s.Total != 1 {
			t.Fatalf("Total = %d", s.Total)
		}
		if s.P50 != 123*time.Millisecond {
			t.Fatalf("P50 = %s", s.P50)
		}
		if s.P90 != 123*time.Millisecond {
			t.Fatalf("P90 = %s", s.P90)
		}
		if s.P99 != 123*time.Millisecond {
			t.Fatalf("P99 = %s", s.P99)
		}
		if s.SuccessRate != 100 {
			t.Fatalf("SuccessRate = %d", s.SuccessRate)
		}
	})

	t.Run("two results", func(t *testing.T) {
		results := []model.Result{
			{Status: 200, Latency: 100 * time.Millisecond},
			{Status: 200, Latency: 200 * time.Millisecond},
		}
		s := Compute(results, time.Second)
		if s.Total != 2 {
			t.Fatalf("Total = %d", s.Total)
		}
		if s.P50 != 100*time.Millisecond {
			t.Fatalf("P50 = %s (nearest-rank should pick lower)", s.P50)
		}
		if s.P90 != 200*time.Millisecond {
			t.Fatalf("P90 = %s", s.P90)
		}
		if s.P99 != 200*time.Millisecond {
			t.Fatalf("P99 = %s", s.P99)
		}
	})

	t.Run("all errors", func(t *testing.T) {
		results := []model.Result{
			{Status: 500, Latency: 50 * time.Millisecond},
			{Status: 0, Latency: 10 * time.Millisecond},
		}
		s := Compute(results, time.Second)
		if s.Total != 2 {
			t.Fatalf("Total = %d", s.Total)
		}
		if s.Successes != 0 {
			t.Fatalf("Successes = %d", s.Successes)
		}
		if s.SuccessRate != 0 {
			t.Fatalf("SuccessRate = %d", s.SuccessRate)
		}
	})

	t.Run("zero elapsed", func(t *testing.T) {
		results := []model.Result{
			{Status: 200, Latency: 100 * time.Millisecond},
		}
		s := Compute(results, 0)
		if s.RequestsPerS != 0 {
			t.Fatalf("RequestsPerS = %f (expected 0)", s.RequestsPerS)
		}
	})
}

func TestComputePercentiles(t *testing.T) {
	results := make([]model.Result, 100)
	for i := range results {
		results[i] = model.Result{Status: 200, Latency: time.Duration(i) * time.Millisecond}
	}
	summary := Compute(results, time.Second)
	if summary.P50 != 49*time.Millisecond {
		t.Fatalf("P50 = %s (expected 49ms)", summary.P50)
	}
	if summary.P90 != 89*time.Millisecond {
		t.Fatalf("P90 = %s (expected 89ms)", summary.P90)
	}
	if summary.P99 != 98*time.Millisecond {
		t.Fatalf("P99 = %s (expected 98ms)", summary.P99)
	}
}

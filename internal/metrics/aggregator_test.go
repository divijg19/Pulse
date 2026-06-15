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
	if summary.MaxLatency != 300*time.Millisecond {
		t.Fatalf("MaxLatency = %s", summary.MaxLatency)
	}
	if summary.RequestsPerS != 2 {
		t.Fatalf("RequestsPerS = %f", summary.RequestsPerS)
	}
}

func TestComputeEmptySummary(t *testing.T) {
	summary := Compute(nil, time.Second)
	if summary.Total != 0 || summary.SuccessRate != 0 || summary.RequestsPerS != 0 {
		t.Fatalf("unexpected empty summary: %+v", summary)
	}
}

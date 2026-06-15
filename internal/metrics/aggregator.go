package metrics

import (
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

type Summary struct {
	Total        int
	Successes    int
	SuccessRate  int
	Average      time.Duration
	MaxLatency   time.Duration
	RequestsPerS float64
}

func Compute(results []model.Result, elapsed time.Duration) Summary {
	summary := Summary{Total: len(results)}
	if len(results) == 0 {
		return summary
	}

	var totalLatency time.Duration
	for _, result := range results {
		if result.Status >= 200 && result.Status < 400 {
			summary.Successes++
		}
		totalLatency += result.Latency
		if result.Latency > summary.MaxLatency {
			summary.MaxLatency = result.Latency
		}
	}

	summary.SuccessRate = int(float64(summary.Successes)/float64(summary.Total)*100 + 0.5)
	summary.Average = totalLatency / time.Duration(summary.Total)
	if elapsed > 0 {
		summary.RequestsPerS = float64(summary.Total) / elapsed.Seconds()
	}

	return summary
}

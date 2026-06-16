package metrics

import (
	"sort"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

type Summary struct {
	Total        int
	Successes    int
	SuccessRate  int
	Average      time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	P50          time.Duration
	P90          time.Duration
	P99          time.Duration
	RequestsPerS float64
}

func Compute(results []model.Result, elapsed time.Duration) Summary {
	summary := Summary{Total: len(results)}
	if len(results) == 0 {
		return summary
	}

	summary.MinLatency = results[0].Latency
	sorted := make([]time.Duration, len(results))
	var totalLatency time.Duration
	for i, result := range results {
		if result.Status >= 200 && result.Status < 400 {
			summary.Successes++
		}
		totalLatency += result.Latency
		if result.Latency > summary.MaxLatency {
			summary.MaxLatency = result.Latency
		}
		if result.Latency < summary.MinLatency {
			summary.MinLatency = result.Latency
		}
		sorted[i] = result.Latency
	}

	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	p := func(percent int) time.Duration {
		rank := (percent*n + 99) / 100
		if rank > n {
			rank = n
		}
		if rank < 1 {
			rank = 1
		}
		return sorted[rank-1]
	}
	summary.P50 = p(50)
	summary.P90 = p(90)
	summary.P99 = p(99)

	summary.SuccessRate = int(float64(summary.Successes)/float64(summary.Total)*100 + 0.5)
	summary.Average = totalLatency / time.Duration(summary.Total)
	if elapsed > 0 {
		summary.RequestsPerS = float64(summary.Total) / elapsed.Seconds()
	}

	return summary
}

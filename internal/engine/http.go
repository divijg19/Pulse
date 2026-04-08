package engine

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

const maxResponseDrainBytes int64 = 1 << 20

var requestClient = &http.Client{Timeout: 30 * time.Second}

func ExecuteSingle(ctx context.Context, url string, method string) model.Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	resp, err := requestClient.Do(req)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseDrainBytes))
		_ = resp.Body.Close()
	}()

	return model.Result{
		Status:    resp.StatusCode,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

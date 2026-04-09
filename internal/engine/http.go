package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

const maxResponseBodyBytes int64 = 10 * 1024

var requestClient = &http.Client{Timeout: 30 * time.Second}

func ExecuteSingle(ctx context.Context, url string, method string, headers map[string]string, body string) model.Result {
	start := time.Now()

	var requestBody io.Reader
	if body != "" {
		requestBody = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	for key, value := range headers {
		if key == "" {
			continue
		}
		req.Header.Add(key, value)
	}

	resp, err := requestClient.Do(req)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if err != nil {
		return model.Result{
			Status:          resp.StatusCode,
			Latency:         time.Since(start),
			Timestamp:       time.Now(),
			Error:           err.Error(),
			ResponseHeaders: flattenHeaders(resp.Header),
		}
	}

	return model.Result{
		Status:          resp.StatusCode,
		Latency:         time.Since(start),
		Timestamp:       time.Now(),
		ResponseHeaders: flattenHeaders(resp.Header),
		ResponseBody:    string(bodyBytes),
	}
}

func flattenHeaders(headers http.Header) map[string]string {
	flattened := make(map[string]string, len(headers))
	for key, values := range headers {
		flattened[key] = strings.Join(values, ", ")
	}
	return flattened
}

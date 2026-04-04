package engine

import (
	"net/http"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

func ExecuteSingle(url string, method string) model.Result {
	start := time.Now()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.Result{ID: "", Status: 0, Latency: 0, Timestamp: time.Now(), Error: err.Error()}
	}
	defer resp.Body.Close()

	return model.Result{
		Status:    resp.StatusCode,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}
}

package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

func TestExport_WritesFile(t *testing.T) {
	dir := t.TempDir()
	results := []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/a"},
		{Status: 404, Latency: 20 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/b"},
	}

	path, err := Export(results, dir)
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	if path == "" {
		t.Fatal("Export returned empty path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("exported file unreadable: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("exported file is empty")
	}
	if got := filepath.Dir(path); got != dir {
		t.Fatalf("expected dir %q, got %q", dir, got)
	}

	// Round-trip: the written JSON must deserialize back to the same results.
	var payload struct {
		Count   int            `json:"count"`
		Results []model.Result `json:"results"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("exported JSON invalid: %v", err)
	}
	if payload.Count != len(results) {
		t.Fatalf("count = %d, want %d", payload.Count, len(results))
	}
	if len(payload.Results) != len(results) {
		t.Fatalf("results length = %d, want %d", len(payload.Results), len(results))
	}
	for i := range results {
		if payload.Results[i].Status != results[i].Status ||
			payload.Results[i].RequestURL != results[i].RequestURL ||
			payload.Results[i].Latency != results[i].Latency {
			t.Fatalf("result %d not preserved: got %+v, want %+v", i, payload.Results[i], results[i])
		}
	}
}

func TestExport_EmptyResults(t *testing.T) {
	if _, err := Export(nil, t.TempDir()); err == nil {
		t.Fatal("Export should fail on empty results")
	}
}

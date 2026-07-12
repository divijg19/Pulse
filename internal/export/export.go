package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

// Export writes the captured results to a timestamped JSON file in dir (or the
// current working directory when dir is empty) and returns the written path.
// Results are serialized with their natural JSON tags so the file round-trips
// with the in-memory model.
func Export(results []model.Result, dir string) (string, error) {
	if len(results) == 0 {
		return "", fmt.Errorf("no results to export")
	}

	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("pulse-results-%s.json", ts)
	path := name
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
		path = filepath.Join(dir, name)
	}

	payload := struct {
		ExportedAt time.Time      `json:"exportedAt"`
		Count      int            `json:"count"`
		Results    []model.Result `json:"results"`
	}{
		ExportedAt: time.Now(),
		Count:      len(results),
		Results:    results,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

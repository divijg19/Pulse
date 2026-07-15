package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

type AuditSurface struct {
	Name  string
	Setup func() Model
}

var AuditSizes = []struct{ W, H int }{
	{80, 24},
	{100, 30},
	{120, 40},
	{160, 40},
}

func AllAuditSurfaces() []AuditSurface {
	return []AuditSurface{
		{"Ready", newReadyModel},
		{"TimelineRunning", newTimelineRunningModel},
		{"LogsRunning", newLogsRunningModel},
		{"TimelineRunningEmpty", newTimelineRunningEmptyModel},
		{"Inspect", newInspectModel},
		{"RequestPayload", newRequestPayloadModel},
		{"Request", newRequestModel},
		{"RequestExec", newRequestExecModel},
		{"ConfirmQuit", newConfirmQuitModel},
	}
}

func WriteAuditCapture(surface AuditSurface, w, h int, dir string) (string, error) {
	m := surface.Setup()
	m.shell.Resize(w, h)
	out := m.View().Content
	name := fmt.Sprintf("%s_%dx%d.ansi", surface.Name, w, h)
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(out), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func newReadyModel() Model {
	return NewModel()
}

func newTimelineRunningModel() Model {
	m := NewModel()
	m.running = true
	m.startedAt = time.Now().Add(-5 * time.Second)
	m.elapsed = 5 * time.Second
	m.results = testResults(20)
	m.selected = 5
	m.workspace.view = TimelineView
	return m
}

func newLogsRunningModel() Model {
	m := newTimelineRunningModel()
	m.workspace.view = LogsView
	return m
}

func newTimelineRunningEmptyModel() Model {
	m := NewModel()
	m.running = true
	m.startedAt = time.Now().Add(-2 * time.Second)
	m.elapsed = 2 * time.Second
	m.workspace.view = TimelineView
	return m
}

func newInspectModel() Model {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = testResults(20)
	m.selected = 3
	return m
}

func newRequestPayloadModel() Model {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.headers = []headerRow{newHeaderRow(), newHeaderRow()}
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")
	m.headers[1].Key.SetValue("Authorization")
	m.headers[1].Value.SetValue("Bearer tok-f8x92k")
	m.bodyInput.SetValue(`{"name":"pulse","version":"1.0.0"}`)
	return m
}

func newRequestModel() Model {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.SetValue("https://httpbin.org/delay/1")
	m.urlInput.Focus()
	return m
}

func newRequestExecModel() Model {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.concurrencyInput.SetValue("10")
	m.concurrencyInput.Focus()
	return m
}

func newConfirmQuitModel() Model {
	m := newTimelineRunningModel()
	m.workspace.dialog = dialogConfirmQuit
	return m
}

func testResults(n int) []model.Result {
	results := make([]model.Result, n)
	statuses := []int{200, 200, 200, 200, 304, 200, 404, 200, 200, 500}
	methods := []string{"GET", "GET", "GET", "POST", "GET", "GET", "GET", "GET", "GET", "GET"}
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://api.example.com/users",
		"https://httpbin.org/status/304",
		"https://api.example.com/orders",
		"https://httpbin.org/redirect/3",
		"https://api.example.com/products",
		"https://httpbin.org/status/404",
		"https://api.example.com/search?q=pulse",
		"https://httpbin.org/delay/2",
		"https://api.example.com/error",
	}
	latencies := []time.Duration{
		87 * time.Millisecond,
		145 * time.Millisecond,
		32 * time.Millisecond,
		210 * time.Millisecond,
		95 * time.Millisecond,
		178 * time.Millisecond,
		12 * time.Millisecond,
		320 * time.Millisecond,
		55 * time.Millisecond,
		450 * time.Millisecond,
	}
	headers := map[string]string{"Content-Type": "application/json"}
	body := `{"ok":true}`

	for i := range results {
		results[i] = model.Result{
			Status:          statuses[i%len(statuses)],
			Latency:         latencies[i%len(latencies)],
			Timestamp:       time.Now().Add(-time.Duration(n-i) * 200 * time.Millisecond),
			RequestMethod:   methods[i%len(methods)],
			RequestURL:      urls[i%len(urls)],
			ResponseHeaders: headers,
			ResponseBody:    body,
		}
	}
	return results
}

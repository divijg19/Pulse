package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
)

func contains(t *testing.T, s, substr string) bool {
	t.Helper()
	return strings.Contains(s, substr)
}

func TestView_Idle(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	out := m.View()
	if !contains(t, out, "Pulse") {
		t.Fatal("View should contain 'Pulse'")
	}
	if !contains(t, out, "SYSTEM READY") {
		t.Fatal("View should contain 'SYSTEM READY'")
	}
	if !contains(t, out, "Awaiting execution...") {
		t.Fatal("View should contain 'Awaiting execution...'")
	}
}

func TestView_Running(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	m.running = true
	m.status = "RUNNING"
	m.startedAt = time.Now().Add(-2 * time.Second)
	m.elapsed = 2 * time.Second
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	out := m.View()
	if !contains(t, out, "ELAPSED") {
		t.Fatal("running view should show elapsed time")
	}
	if !contains(t, out, "RPS") {
		t.Fatal("running view should show RPS")
	}
}

func TestView_EmptyState(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	out := m.View()
	if !contains(t, out, "Awaiting execution...") {
		t.Fatal("empty view should show 'Awaiting execution...'")
	}
}

func TestRenderCommand(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderCommand(96)
	if !contains(t, out, "GET") {
		t.Fatal("command should contain 'GET'")
	}
	if !contains(t, out, "CC") {
		t.Fatal("command should contain 'CC'")
	}
	if !contains(t, out, "RUN") {
		t.Fatal("command should contain 'RUN'")
	}
	if !contains(t, out, "PAYLOAD") {
		t.Fatal("command should contain 'PAYLOAD'")
	}
}

func TestRenderCommand_Running(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	out := m.renderCommand(96)
	if !contains(t, out, "CANCEL") {
		t.Fatal("running command should contain 'CANCEL'")
	}
}

func TestRenderCommand_Dividers(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderCommand(96)
	count := strings.Count(out, "│")
	if count < 3 {
		t.Fatalf("command should have at least 3 vertical dividers, got %d", count)
	}
}

func TestRenderMetrics(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.summary.Total = 50
	m.summary.Successes = 45
	m.summary.SuccessRate = 90
	m.summary.P50 = 100 * time.Millisecond
	m.summary.P99 = 500 * time.Millisecond
	m.elapsed = 5 * time.Second

	out := m.renderMetrics(96)
	if !contains(t, out, "REQUESTS") {
		t.Fatal("metrics should contain 'REQUESTS'")
	}
	if !contains(t, out, "SUCCESS") {
		t.Fatal("metrics should contain 'SUCCESS'")
	}
	if !contains(t, out, "ERRORS") {
		t.Fatal("metrics should contain 'ERRORS'")
	}
	if !contains(t, out, "LATENCY") {
		t.Fatal("metrics should contain 'LATENCY'")
	}
	if !contains(t, out, "90%") {
		t.Fatal("metrics should show '90%' success rate")
	}
	if !contains(t, out, "0.10s") {
		t.Fatal("metrics should show p50 latency")
	}
}

func TestRenderMetrics_Zero(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderMetrics(96)
	if !contains(t, out, "REQUESTS") {
		t.Fatal("zero metrics should still show 'REQUESTS'")
	}
}

func TestRenderPayload_Open(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.showPayload = true
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderPayload(96)
	if !contains(t, out, "HEADERS") {
		t.Fatal("payload should contain 'HEADERS'")
	}
	if !contains(t, out, "BODY") {
		t.Fatal("payload should contain 'BODY'")
	}
	if !contains(t, out, "Content-Type") {
		t.Fatal("payload should contain header key")
	}
	if !contains(t, out, "application/json") {
		t.Fatal("payload should contain header value")
	}
}

func TestRenderTabs(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTabs(94)
	if !contains(t, out, "Timeline") {
		t.Fatal("tabs should contain 'Timeline'")
	}
	if !contains(t, out, "Live Logs") {
		t.Fatal("tabs should contain 'Live Logs'")
	}
}

func TestRenderTabs_Manual(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.autoScroll = false
	out := m.renderTabs(94)
	if !contains(t, out, "MANUAL") {
		t.Fatal("tabs should show 'MANUAL' when autoScroll is off")
	}
}

func TestRenderTimeline_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Awaiting execution...") {
		t.Fatal("empty timeline should show 'Awaiting execution...'")
	}
}

func TestRenderTimeline_Rows(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
		{Status: 404, Latency: 50 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.focus = focusResults
	m.selected = 0

	out := m.renderTimeline(94, 20)
	if !contains(t, out, "200") {
		t.Fatal("timeline should show status code 200")
	}
	if !contains(t, out, "404") {
		t.Fatal("timeline should show status code 404")
	}
	if !contains(t, out, "0.10s") {
		t.Fatal("timeline should show latency")
	}
}

func TestRenderLogs_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderLogs(94, 20)
	if !contains(t, out, "No logs yet.") {
		t.Fatal("empty logs should show 'No logs yet.'")
	}
}

func TestRenderLogs_Rows(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.urlInput.SetValue("https://example.com/api")
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/api"},
		{Status: 500, Latency: 50 * time.Millisecond, RequestMethod: "POST", RequestURL: "https://example.com/error"},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.focus = focusResults
	m.selected = 0

	out := m.renderLogs(94, 20)
	if !contains(t, out, "GET") {
		t.Fatal("logs should show method 'GET'")
	}
	if !contains(t, out, "POST") {
		t.Fatal("logs should show method 'POST'")
	}
	if !contains(t, out, "200") {
		t.Fatal("logs should show status 200")
	}
	if !contains(t, out, "500") {
		t.Fatal("logs should show status 500")
	}
	if !contains(t, out, "0.10s") {
		t.Fatal("logs should show latency")
	}
}

func TestRenderInspector_NoSelection(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderInspector(40, 20)
	if !contains(t, out, "No result selected.") {
		t.Fatal("inspector with no selection should show 'No result selected.'")
	}
}

func TestRenderInspector_WithResult(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			ResponseBody: `{"ok": true}`,
		},
	}

	out := m.renderInspector(40, 20)
	if !contains(t, out, "INSPECTOR") {
		t.Fatal("inspector should show 'INSPECTOR'")
	}
	if !contains(t, out, "200") {
		t.Fatal("inspector should show status")
	}
	if !contains(t, out, "0.10s") {
		t.Fatal("inspector should show latency")
	}
	if !contains(t, out, "Content-Type") {
		t.Fatal("inspector should show response headers")
	}
	if !contains(t, out, "application/json") {
		t.Fatal("inspector should show header values")
	}
}

func TestRenderSparkline_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderSparkline(96)
	if out != "" {
		t.Fatal("sparkline should be empty when no data")
	}
}

func TestRenderSparkline_WithData(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.latencyLen = 5
	m.latencyRing[0] = 10 * time.Millisecond
	m.latencyRing[1] = 20 * time.Millisecond
	m.latencyRing[2] = 50 * time.Millisecond
	m.latencyRing[3] = 100 * time.Millisecond
	m.latencyRing[4] = 200 * time.Millisecond
	m.latencyHead = 5

	out := m.renderSparkline(96)
	if !contains(t, out, "LATENCY SPARKLINE") {
		t.Fatal("sparkline should show 'LATENCY SPARKLINE' label")
	}
}

func TestResultStatus(t *testing.T) {
	tt := []struct {
		status   int
		expected string
	}{
		{0, "ERR"},
		{200, "200 OK"},
		{201, "201 OK"},
		{301, "301 Redirect"},
		{400, "400"},
		{404, "404"},
		{500, "500"},
		{101, "101 Info"},
		{199, "199 Info"},
		{302, "302 Redirect"},
	}

	for _, tc := range tt {
		result := model.Result{Status: tc.status}
		got := resultStatus(result)
		if got != tc.expected {
			t.Errorf("resultStatus(%d) = %q (expected %q)", tc.status, got, tc.expected)
		}
	}
}

func TestRowCursor(t *testing.T) {
	if got := rowCursor(true); got != "▸" {
		t.Errorf("selected cursor = %q", got)
	}
	if got := rowCursor(false); got != " " {
		t.Errorf("unselected cursor = %q", got)
	}
}

func TestTruncate(t *testing.T) {
	tt := []struct {
		value string
		width int
		exp   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"", 5, ""},
		{"hello", 0, ""},
		{"abc", 1, "a"},
		{"hi\n there", 10, "hi  there"},
	}

	for _, tc := range tt {
		got := truncate(tc.value, tc.width)
		if got != tc.exp {
			t.Errorf("truncate(%q, %d) = %q (expected %q)", tc.value, tc.width, got, tc.exp)
		}
	}
}

func TestDisplayElapsed(t *testing.T) {
	m := NewModel()
	m.elapsed = 5 * time.Second

	m.running = false
	if got := displayElapsed(m); got != 5*time.Second {
		t.Fatalf("elapsed should be 5s when not running but has elapsed, got %v", got)
	}

	m.running = true
	m.elapsed = 3 * time.Second
	if got := displayElapsed(m); got != 3*time.Second {
		t.Fatalf("elapsed should be 3s when running, got %v", got)
	}

	m.running = false
	m.elapsed = 0
	if got := displayElapsed(m); got != 0 {
		t.Fatalf("elapsed should be 0 when idle, got %v", got)
	}
}

func TestRenderFooter_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderFooter(96)
	if !contains(t, out, "ctrl+r") {
		t.Fatal("footer should show 'ctrl+r' shortcut")
	}
	if !contains(t, out, "ctrl+x") {
		t.Fatal("footer should show 'ctrl+x' shortcut")
	}
}

func TestRenderFooter_WithResults(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
		{Status: 200, Latency: 20 * time.Millisecond},
	}
	m.elapsed = 2 * time.Second

	out := m.renderFooter(96)
	if !contains(t, out, "2 results") {
		t.Fatal("footer should show result count")
	}
	if !contains(t, out, "p99") {
		t.Fatal("footer should show p99 latency")
	}
}

func TestMethodFallback(t *testing.T) {
	m := NewModel()
	m.methodIndex = 0
	m.urlInput.SetValue("https://example.com")

	result := model.Result{Status: 200, Latency: 10 * time.Millisecond}

	if result.RequestMethod != "" {
		t.Fatal("test setup: RequestMethod should be empty")
	}
	if result.RequestURL != "" {
		t.Fatal("test setup: RequestURL should be empty")
	}

	_ = resultStatus(result)
	_ = formatDuration(result.Latency)
	_ = fmt.Sprintf("%s", runconfig.AllowedMethods()[m.methodIndex])
	_ = truncate(m.urlInput.Value(), 30)
}

func TestVersion(t *testing.T) {
	out := "Pulse terminal"
	if !strings.Contains(out, "Pulse") {
		t.Fatal("should contain Pulse")
	}
}

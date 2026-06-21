package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
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
	if !contains(t, out, "GET") {
		t.Fatal("View should contain method")
	}
	if !contains(t, out, "Ctrl+R to run") {
		t.Fatal("View should contain 'Ctrl+R to run' in Ready surface")
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
	if !contains(t, out, "RUNNING") {
		t.Fatal("running view should show running indicator")
	}
	if !contains(t, out, "r/s") {
		t.Fatal("running view should show requests per second")
	}
	if !contains(t, out, "Timeline") {
		t.Fatal("running view should show Timeline identity")
	}
}

func TestRenderReady(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	out := m.renderReady(100, 26)
	if !contains(t, out, "httpbin") {
		t.Fatal("Ready should show URL")
	}
	if !contains(t, out, "CC 10") {
		t.Fatal("Ready should show concurrency")
	}
	if !contains(t, out, "Ctrl+R to run") {
		t.Fatal("Ready should show run CTA")
	}
	if !contains(t, out, "Endpoint") {
		t.Fatal("Ready should show endpoint action link")
	}
	if !contains(t, out, "Concurrency") {
		t.Fatal("Ready should show concurrency action link")
	}
	if !contains(t, out, "Payload") {
		t.Fatal("Ready should show payload action link")
	}
}

func TestRenderReady_HidesAfterFirstRun(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30

	out := m.View()
	if !contains(t, out, "Ctrl+R to run") {
		t.Fatal("first launch should show Ready surface")
	}

	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out = m.View()
	if contains(t, out, "Ready") {
		t.Fatal("after results exist, Ready should not appear")
	}
}

func TestRenderTopBar_ShowsMethodAndURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTopBar(100)
	if !contains(t, out, "GET") {
		t.Fatal("top bar should contain method")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("top bar should contain URL")
	}
}

func TestRenderTopBar_ShowsCC(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTopBar(100)
	if !contains(t, out, "CC") {
		t.Fatal("top bar should show CC")
	}
}

func TestRenderTopBar_QueryTruncation(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.urlInput.SetValue("https://api.example.com/users?page=1")
	out := m.renderTopBar(100)
	if contains(t, out, "?page=1") {
		t.Fatal("top bar should truncate query string from URL")
	}
	if !contains(t, out, "api.example.com/users") {
		t.Fatal("top bar should show truncated URL without query")
	}
}

func TestMetricsString(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.summary.Total = 50
	m.summary.Successes = 45
	m.summary.SuccessRate = 90
	m.summary.P90 = 100 * time.Millisecond
	m.summary.P99 = 500 * time.Millisecond
	m.summary.MaxLatency = 500 * time.Millisecond
	m.elapsed = 5 * time.Second

	out := m.metricsString()
	if !contains(t, out, "90% ok") {
		t.Fatal("metrics should show success rate")
	}
	if !contains(t, out, "r/s") {
		t.Fatal("metrics should show requests per second")
	}
	if !contains(t, out, "p90") {
		t.Fatal("metrics should show p90 latency")
	}
	if !contains(t, out, "p99") {
		t.Fatal("metrics should show p99 latency")
	}
}

func TestMetricsString_Zero(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.metricsString()
	if out != "" {
		t.Fatal("idle metrics with no results should be empty")
	}
}

func TestMetricsString_HiddenWhenIdle(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.summary.Total = 0
	m.summary.SuccessRate = 0
	m.elapsed = 0

	out := m.metricsString()
	if out != "" {
		t.Fatal("metrics should be hidden when idle with no results")
	}

	m.running = true
	out = m.metricsString()
	if out == "" {
		t.Fatal("metrics should appear when running even with no results")
	}
}

func TestMetricsString_Values(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.summary.Total = 10
	m.summary.Successes = 10
	m.summary.SuccessRate = 100
	m.elapsed = 5 * time.Second

	out := m.metricsString()
	if !contains(t, out, "100% ok") {
		t.Fatal("100% success rate should show '100% ok'")
	}

	m.summary.Successes = 9
	m.summary.SuccessRate = 90
	out = m.metricsString()
	if !contains(t, out, "90% ok") {
		t.Fatal("90% success rate should show '90% ok'")
	}
}

func TestMetricsString_RunningRPS(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.elapsed = 2 * time.Second
	m.summary.Total = 100
	m.summary.Successes = 95
	m.summary.SuccessRate = 95
	m.summary.P90 = 100 * time.Millisecond
	m.summary.P99 = 500 * time.Millisecond
	m.summary.MaxLatency = 500 * time.Millisecond

	out := m.metricsString()
	if !contains(t, out, "95% ok") {
		t.Fatal("metrics should show success rate")
	}
}

func TestMetricsString_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.elapsed = 100 * time.Millisecond
	m.summary.Total = 0
	m.summary.SuccessRate = 0

	out := m.metricsString()
	if out == "" {
		t.Fatal("metrics should appear when running, even with zero results")
	}
	if !contains(t, out, "% ok") {
		t.Fatal("running empty metrics should show success rate")
	}
	if !contains(t, out, "r/s") {
		t.Fatal("running empty metrics should show r/s")
	}
}

func TestRenderStatusBar_Normal(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderStatusBar(100)
	if !contains(t, out, "OBSERVE") {
		t.Fatal("status bar should show 'OBSERVE' mode")
	}
}

func TestRenderStatusBar_Running(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	out := m.renderStatusBar(100)
	if !contains(t, out, "RUNNING") {
		t.Fatal("status bar should show 'RUNNING' mode")
	}
}

func TestRenderStatusBar_EndpointDialog(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogEndpoint
	out := m.renderStatusBar(100)
	if !contains(t, out, "ENDPOINT") {
		t.Fatal("status bar should show 'ENDPOINT' mode")
	}
}

func TestRenderStatusBar_CCDialog(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogConcurrency
	out := m.renderStatusBar(100)
	if !contains(t, out, "CONCURRENCY") {
		t.Fatal("status bar should show 'CONCURRENCY' mode")
	}
}

func TestRenderStatusBar_Inspecting(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.mode = modeInspect
	out := m.renderStatusBar(100)
	if !contains(t, out, "INSPECTING") {
		t.Fatal("status bar should show 'INSPECTING' mode")
	}
}

func TestRenderStatusBar_QuitConfirm(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogConfirmQuit
	out := m.renderStatusBar(100)
	if !contains(t, out, "quit") {
		t.Fatal("status bar should show quit confirmation")
	}
}

func TestRenderTimeline_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Timeline") {
		t.Fatal("timeline should show identity header")
	}
}

func TestRenderTimeline_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Timeline") {
		t.Fatal("empty timeline should show Timeline identity")
	}
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("empty timeline should show '▶  Ctrl+R to run'")
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

func TestRenderLogs_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/api"},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderLogs(94, 20)
	if !contains(t, out, "Logs") {
		t.Fatal("logs should show identity header")
	}
}

func TestRenderLogs_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderLogs(94, 20)
	if !contains(t, out, "Logs") {
		t.Fatal("empty logs should show Logs identity")
	}
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("empty logs should show '▶  Ctrl+R to run'")
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

func TestRenderInspect_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.selected = 0
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}

	out := m.renderInspect(40, 20)
	if !contains(t, out, "Inspector") {
		t.Fatal("inspector should show identity header")
	}
	if !contains(t, out, "Result #1") {
		t.Fatal("inspector should show result number")
	}
}

func TestRenderInspect_NoSelection(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderInspect(40, 20)
	if !contains(t, out, "No result selected.") {
		t.Fatal("inspector with no selection should show 'No result selected.'")
	}
}

func TestRenderInspect_WithResult(t *testing.T) {
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

	out := m.renderInspect(40, 20)
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

func TestRenderInspect_NoMetrics(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.selected = 0
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.Total = 50
	m.summary.SuccessRate = 90
	m.elapsed = 5 * time.Second

	out := m.renderInspect(40, 20)
	if contains(t, out, "% ok") {
		t.Fatal("inspector should NOT show aggregate metrics")
	}
	if contains(t, out, "r/s") {
		t.Fatal("inspector should NOT show requests per second")
	}
}

func TestRenderInspect_WithError(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  500,
			Latency: 100 * time.Millisecond,
			Error:   "connection refused",
		},
	}

	out := m.renderInspect(40, 20)
	if !contains(t, out, "connection refused") {
		t.Fatal("inspector should show error message")
	}
	if !contains(t, out, "Error:") {
		t.Fatal("inspector should show 'Error:' prefix")
	}
}

func TestRenderInspect_NoHeaders(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
		},
	}

	out := m.renderInspect(40, 20)
	if !contains(t, out, "No headers captured.") {
		t.Fatal("inspector should show 'No headers captured.' when no response headers")
	}
}

func TestRenderInspect_NoBody(t *testing.T) {
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
		},
	}

	out := m.renderInspect(40, 20)
	if !contains(t, out, "No body captured.") {
		t.Fatal("inspector should show 'No body captured.' when no response body")
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
		{50, "50"},
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
	if got := rowCursor(true); got != "▶" {
		t.Errorf("selected cursor = %q", got)
	}
	if got := rowCursor(false); got != " " {
		t.Errorf("unselected cursor = %q", got)
	}
}

func TestTruncateURL(t *testing.T) {
	tt := []struct {
		raw   string
		width int
		exp   string
	}{
		{"https://api.example.com/v1/users", 100, "api.example.com/v1/users"},
		{"https://api.example.com/v1/users?page=1", 100, "api.example.com/v1/users"},
		{"http://example.com/posts", 100, "example.com/posts"},
		{"https://x.com/a", 100, "x.com/a"},
		{"", 100, ""},
	}
	for _, tc := range tt {
		got := truncateURL(tc.raw, tc.width)
		if got != tc.exp {
			t.Errorf("truncateURL(%q, %d) = %q (expected %q)", tc.raw, tc.width, got, tc.exp)
		}
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

func TestRenderTimeline_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.status = "RUNNING"

	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Timeline") {
		t.Fatal("running empty timeline should show Timeline identity")
	}
	if !contains(t, out, "⏳  Waiting for results...") {
		t.Fatal("running empty timeline should show '⏳  Waiting for results...'")
	}
}

func TestRenderTimeline_IdleNoURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.urlInput.SetValue("")

	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Timeline") {
		t.Fatal("idle empty timeline should show Timeline identity")
	}
	if !contains(t, out, "Enter a URL to begin") {
		t.Fatal("idle empty timeline with no URL should show 'Enter a URL to begin'")
	}
}

func TestRenderTimeline_IdleWithURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.urlInput.SetValue("https://example.com/api")

	out := m.renderTimeline(94, 20)
	if !contains(t, out, "Timeline") {
		t.Fatal("idle empty timeline should show Timeline identity")
	}
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("idle empty timeline with URL should show '▶  Ctrl+R to run'")
	}
}

func TestRenderLogs_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.status = "RUNNING"

	out := m.renderLogs(94, 20)
	if !contains(t, out, "Logs") {
		t.Fatal("running empty logs should show Logs identity")
	}
	if !contains(t, out, "📭  No results yet...") {
		t.Fatal("running empty logs should show '📭  No results yet...'")
	}
}

func TestRenderLogs_IdleNoURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.urlInput.SetValue("")

	out := m.renderLogs(94, 20)
	if !contains(t, out, "Logs") {
		t.Fatal("idle empty logs should show Logs identity")
	}
	if !contains(t, out, "Enter a URL to begin") {
		t.Fatal("idle empty logs with no URL should show 'Enter a URL to begin'")
	}
}

func TestRenderLogs_IdleWithURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.urlInput.SetValue("https://example.com/api")

	out := m.renderLogs(94, 20)
	if !contains(t, out, "Logs") {
		t.Fatal("idle empty logs should show Logs identity")
	}
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("idle empty logs with URL should show '▶  Ctrl+R to run'")
	}
}

func TestView_WidthMinClamp(t *testing.T) {
	m := NewModel()
	m.width = 40
	m.height = 30

	out := m.View()
	if !contains(t, out, "GET") {
		t.Fatal("View should contain method even at small width")
	}
}

func TestView_PayloadNotShown(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30

	out := m.View()
	if contains(t, out, "HEADERS") {
		t.Fatal("View should not contain HEADERS when payload dialog is closed")
	}
	if contains(t, out, "BODY") {
		t.Fatal("View should not contain BODY when payload dialog is closed")
	}
}

func TestRenderPayload_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogPayload
	m.selectedHead = 0
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderPayload(96)
	if !contains(t, out, "Payload") {
		t.Fatal("payload should show identity header")
	}
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

func TestRenderPayload_NoHeadersConfigured(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogPayload
	m.selectedHead = 0

	out := m.renderPayload(96)
	if !contains(t, out, "Payload") {
		t.Fatal("payload should show identity header")
	}
	if !contains(t, out, "No headers configured.") {
		t.Fatal("payload should show 'No headers configured.' when no headers")
	}
}

func TestRenderPayload_EmptyBodyPlaceholder(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.dialog = dialogPayload
	m.selectedHead = 0
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderPayload(96)
	if !contains(t, out, "Payload") {
		t.Fatal("payload should show identity header")
	}
	if !contains(t, out, `{"name":"pulse"}`) {
		t.Fatal("payload should show body placeholder when body is empty")
	}
}

func TestRenderEndpoint_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderEndpoint(100)
	if !contains(t, out, "Endpoint") {
		t.Fatal("endpoint should show identity header")
	}
	if !contains(t, out, "URL") {
		t.Fatal("endpoint editor should show URL label")
	}
	if !contains(t, out, "Method") {
		t.Fatal("endpoint editor should show Method label")
	}
	if !contains(t, out, "GET") {
		t.Fatal("endpoint editor should show method options")
	}
}

func TestRenderConcurrency_Identity(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.setConcurrency(7)
	out := m.renderConcurrency(100)
	if !contains(t, out, "Concurrency") {
		t.Fatal("concurrency should show identity header")
	}
	if !contains(t, out, "7") {
		t.Fatal("concurrency editor should show current value")
	}
	if !contains(t, out, "1–100") {
		t.Fatal("concurrency should show range affordance")
	}
}

func TestRenderWorkspace_InspectorDrillDown(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	m.mode = modeInspect
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{"Content-Type": "application/json"},
			ResponseBody:    `{"ok": true}`},
	}
	m.selected = 0

	out := m.View()
	if !contains(t, out, "Inspector") {
		t.Fatal("workspace with inspector should show Inspector header")
	}
	if !contains(t, out, "Result #1") {
		t.Fatal("workspace with inspector should show result number")
	}
	if !contains(t, out, "200") {
		t.Fatal("workspace should show result status")
	}
}

func TestRenderTimeline_Rows_Selected(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	row := m.renderTimelineRow(0, m.results[0], m.summary.MaxLatency, 94, true)
	if !contains(t, row, "200") {
		t.Fatal("selected row should show status")
	}
	if !contains(t, row, "▶") {
		t.Fatal("selected row should show cursor")
	}
}

func TestVisibleWindow(t *testing.T) {
	tt := []struct {
		total    int
		selected int
		height   int
		expected int
	}{
		{0, 0, 10, 0},
		{5, 2, 10, 0},
		{10, 5, 5, 3},
		{10, 9, 5, 5},
		{10, 0, 5, 0},
	}
	for _, tc := range tt {
		got := visibleWindow(tc.total, tc.selected, tc.height)
		if got != tc.expected {
			t.Errorf("visibleWindow(%d,%d,%d) = %d (expected %d)", tc.total, tc.selected, tc.height, got, tc.expected)
		}
	}
}

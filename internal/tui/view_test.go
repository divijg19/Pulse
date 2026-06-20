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
	if !contains(t, out, "IDLE") {
		t.Fatal("View should contain 'IDLE'")
	}
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("View should contain '▶  Ctrl+R to run'")
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
}

func TestRenderTopBar_Normal(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTopBar(100)
	if !contains(t, out, "GET") {
		t.Fatal("top bar should contain method")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("top bar should contain URL")
	}
	if !contains(t, out, "IDLE") {
		t.Fatal("top bar should contain state")
	}
}

func TestRenderTopBar_EditURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusURL
	out := m.renderTopBar(100)
	if !contains(t, out, "GET") {
		t.Fatal("top bar should contain method")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("top bar should contain URL value when URL is focused")
	}
}

func TestRenderTopBar_Running(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.elapsed = 2 * time.Second
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	out := m.renderTopBar(100)
	if !contains(t, out, "RUNNING") {
		t.Fatal("top bar should show RUNNING state")
	}
	if !contains(t, out, "1 req") {
		t.Fatal("top bar should show request count")
	}
}

func TestRenderTopBar_Completed(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.status = "COMPLETE"
	m.elapsed = 5 * time.Second
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	out := m.renderTopBar(100)
	if !contains(t, out, "COMPLETED") {
		t.Fatal("top bar should show COMPLETED state")
	}
	if !contains(t, out, "1 req") {
		t.Fatal("top bar should show request count on completion")
	}
}

func TestRenderTopBar_Cancelled(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.status = "CANCELLED"
	out := m.renderTopBar(100)
	if !contains(t, out, "CANCELLED") {
		t.Fatal("top bar should show CANCELLED state via renderTopBarStatus")
	}
	if !contains(t, out, "CC") {
		t.Fatal("top bar should show CC when status is CANCELLED")
	}
}

func TestRenderTopBar_StatusError(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.status = "ERROR: connection refused"
	out := m.renderTopBar(100)
	if !contains(t, out, "ERROR") {
		t.Fatal("top bar should show ERROR status via renderTopBarStatus")
	}
}

func TestRenderTopBar_IdleDefault(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.status = "SYSTEM READY"
	out := m.renderTopBar(100)
	if !contains(t, out, "IDLE") {
		t.Fatal("top bar should show IDLE when status is SYSTEM READY")
	}
	if !contains(t, out, "CC") {
		t.Fatal("top bar should show CC when idle")
	}
}

func TestRenderTopBar_Inspecting(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.inspector = true
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.selected = 0
	out := m.renderTopBar(100)
	if !contains(t, out, "INSPECTING") {
		t.Fatal("top bar should show INSPECTING state")
	}
}

func TestRenderTopBar_QuitConfirm(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.confirmQuit = true
	out := m.renderTopBar(100)
	if !contains(t, out, "QUIT?") {
		t.Fatal("top bar should show QUIT? confirmation")
	}
}

func TestRenderTopBar_EditCC(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusConcurrency
	out := m.renderTopBar(100)
	if !contains(t, out, "CC") {
		t.Fatal("edit CC top bar should contain 'CC' label")
	}
	if !contains(t, out, "IDLE") {
		t.Fatal("edit CC top bar should show IDLE state")
	}
}

func TestRenderTopBar_URLFocus(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusURL
	out := m.renderTopBar(100)
	if !contains(t, out, "[httpbin.org/delay/1]") {
		t.Fatal("URL focus top bar should show bracketed URL")
	}
}

func TestRenderTopBar_CCFocus(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusConcurrency
	out := m.renderTopBar(100)
	if !contains(t, out, "[10]") {
		t.Fatal("CC focus top bar should show bracketed concurrency")
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

func TestRenderMetrics(t *testing.T) {
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

	out := m.renderMetrics(100)
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
	if contains(t, out, "p50") {
		t.Fatal("metrics should not show p50 latency")
	}
	if contains(t, out, "req") {
		t.Fatal("metrics should not show request count")
	}
}

func TestRenderMetrics_Zero(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderMetrics(100)
	if out != "" {
		t.Fatal("idle metrics with no results should be empty")
	}
}

func TestRenderMetrics_HiddenWhenIdle(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = false
	m.summary.Total = 0
	m.summary.SuccessRate = 0
	m.elapsed = 0

	out := m.renderMetrics(100)
	if out != "" {
		t.Fatal("metrics should be hidden when idle with no results")
	}

	m.running = true
	out = m.renderMetrics(100)
	if out == "" {
		t.Fatal("metrics should appear when running even with no results")
	}
}

func TestRenderMetrics_ErrorColor(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.summary.Total = 10
	m.summary.Successes = 10
	m.summary.SuccessRate = 100
	m.elapsed = 5 * time.Second

	outNoErrors := m.renderMetrics(100)
	if !contains(t, outNoErrors, "100% ok") {
		t.Fatal("metrics should show 100% ok")
	}

	m.summary.Successes = 7
	m.summary.SuccessRate = 70
	outWithErrors := m.renderMetrics(100)
	if !contains(t, outWithErrors, "70% ok") {
		t.Fatal("metrics with errors should show 70% ok")
	}

	if outNoErrors == outWithErrors {
		t.Fatal("different success rates should produce different output")
	}
}

func TestRenderMetrics_RunningRPS(t *testing.T) {
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

	out := m.renderMetrics(100)
	if !contains(t, out, "95% ok") {
		t.Fatal("metrics should show success rate")
	}
}

func TestRenderMetrics_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.elapsed = 100 * time.Millisecond
	m.summary.Total = 0
	m.summary.SuccessRate = 0

	out := m.renderMetrics(100)
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

func TestRenderTabs(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTabStrip(100)
	if !contains(t, out, "Timeline") {
		t.Fatal("tab strip should contain 'Timeline'")
	}
	if !contains(t, out, "Logs") {
		t.Fatal("tab strip should contain 'Logs'")
	}
}

func TestRenderTabs_NoManualIndicator(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.autoScroll = false
	out := m.renderTabStrip(100)
	if contains(t, out, "MANUAL") {
		t.Fatal("tab strip should not show 'MANUAL' indicator")
	}
	if !contains(t, out, "Timeline") {
		t.Fatal("tab strip should still show tabs when autoScroll is off")
	}
}

func TestRenderTabs_ActiveMarker(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTabStrip(100)
	if !contains(t, out, "▶ Timeline") {
		t.Fatal("tab strip should show ▶ marker on active Timeline tab")
	}
	if !contains(t, out, "  Logs") {
		t.Fatal("tab strip should show inactive Logs tab without marker")
	}

	m.activeTab = tabLogs
	out = m.renderTabStrip(100)
	if !contains(t, out, "▶ Logs") {
		t.Fatal("tab strip should show ▶ marker on active Logs tab")
	}
	if !contains(t, out, "  Timeline") {
		t.Fatal("tab strip should show inactive Timeline tab without marker")
	}
}

func TestRenderStatusBar_Normal(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusResults
	out := m.renderStatusBar(100)
	if !contains(t, out, "NORMAL") {
		t.Fatal("status bar should show 'NORMAL' mode")
	}
	if !contains(t, out, "TAB focus") {
		t.Fatal("status bar should show tab hint")
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

func TestRenderStatusBar_EditURL(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.focus = focusURL
	out := m.renderStatusBar(100)
	if !contains(t, out, "EDIT URL") {
		t.Fatal("status bar should show 'EDIT URL' mode")
	}
}

func TestRenderStatusBar_Inspecting(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.inspector = true
	out := m.renderStatusBar(100)
	if !contains(t, out, "INSPECTING") {
		t.Fatal("status bar should show 'INSPECTING' mode")
	}
}

func TestRenderStatusBar_QuitConfirm(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.confirmQuit = true
	out := m.renderStatusBar(100)
	if !contains(t, out, "PRESS Q AGAIN TO QUIT") {
		t.Fatal("status bar should show quit confirmation")
	}
}

func TestRenderTimeline_Empty(t *testing.T) {
	m := NewModel()
	m.width = 100
	out := m.renderTimeline(94, 20)
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
	if !contains(t, out, "▶  Ctrl+R to run") {
		t.Fatal("idle empty logs with URL should show '▶  Ctrl+R to run'")
	}
}

func TestRenderInspector_WithError(t *testing.T) {
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

	out := m.renderInspector(40, 20)
	if !contains(t, out, "INSPECTOR") {
		t.Fatal("inspector should show 'INSPECTOR'")
	}
	if !contains(t, out, "connection refused") {
		t.Fatal("inspector should show error message")
	}
	if !contains(t, out, "Error:") {
		t.Fatal("inspector should show 'Error:' prefix")
	}
}

func TestRenderInspector_NoHeaders(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
		},
	}

	out := m.renderInspector(40, 20)
	if !contains(t, out, "No headers captured.") {
		t.Fatal("inspector should show 'No headers captured.' when no response headers")
	}
}

func TestRenderInspector_NoBody(t *testing.T) {
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

	out := m.renderInspector(40, 20)
	if !contains(t, out, "No body captured.") {
		t.Fatal("inspector should show 'No body captured.' when no response body")
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
	if !contains(t, out, "IDLE") {
		t.Fatal("View should show state even at small width")
	}
}

func TestView_PayloadNotShown(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	m.showPayload = false

	out := m.View()
	if contains(t, out, "HEADERS") {
		t.Fatal("View should not contain HEADERS when payload is hidden")
	}
	if contains(t, out, "BODY") {
		t.Fatal("View should not contain BODY when payload is hidden")
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

func TestRenderPayload_NoHeadersConfigured(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.showPayload = true

	out := m.renderPayload(96)
	if !contains(t, out, "No headers configured.") {
		t.Fatal("payload should show 'No headers configured.' when no headers")
	}
}

func TestRenderPayload_EmptyBodyPlaceholder(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.showPayload = true
	m.focus = focusHeaders
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderPayload(96)
	if !contains(t, out, `{"name":"pulse"}`) {
		t.Fatal("payload should show body placeholder when body is empty")
	}
}

func TestRenderWorkspace_InspectorStacked(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.height = 30
	m.inspector = true
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{"Content-Type": "application/json"},
			ResponseBody:    `{"ok": true}`},
	}
	m.selected = 0
	m.focus = focusResults

	out := m.View()
	if !contains(t, out, "INSPECTOR") {
		t.Fatal("workspace with inspector should show INSPECTOR")
	}
	if !contains(t, out, "Timeline") {
		t.Fatal("workspace should show Timeline tab")
	}
	if !contains(t, out, "200") {
		t.Fatal("workspace should show result status")
	}
}

func TestRenderMetrics_SuccessColor(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.running = true
	m.summary.Total = 10
	m.summary.Successes = 10
	m.summary.SuccessRate = 100
	m.elapsed = 5 * time.Second

	out100 := m.renderMetrics(100)
	if !contains(t, out100, "100% ok") {
		t.Fatal("100% success rate should show '100% ok'")
	}

	m.summary.Successes = 9
	m.summary.SuccessRate = 90
	out90 := m.renderMetrics(100)
	if !contains(t, out90, "90% ok") {
		t.Fatal("90% success rate should show '90% ok'")
	}

	if out100 == out90 {
		t.Fatal("different success rates should produce different output")
	}
}

func TestRenderTimeline_Rows_Selected(t *testing.T) {
	m := NewModel()
	m.width = 100
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.focus = focusResults
	m.selected = 0

	row := m.renderTimelineRow(0, m.results[0], m.summary.MaxLatency, 94, true)
	if !contains(t, row, "200") {
		t.Fatal("selected row should show status")
	}
	if !contains(t, row, "▶") {
		t.Fatal("selected row should show cursor")
	}
}

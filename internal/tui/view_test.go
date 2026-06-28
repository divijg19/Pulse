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
	m.shell.Resize(100, 30)
	out := m.View()
	if !contains(t, out, "GET") {
		t.Fatal("View should contain method")
	}
	if !contains(t, out, "OBSERVE") {
		t.Fatal("View should contain OBSERVE identity in Ready surface")
	}
}

func TestView_Running(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.running = true
	m.status = "RUNNING"
	m.startedAt = time.Now().Add(-2 * time.Second)
	m.elapsed = 2 * time.Second
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	out := m.View()
	if !contains(t, out, "r/s") {
		t.Fatal("running view should show requests per second")
	}
	if !contains(t, out, "Timeline") {
		t.Fatal("running view should show Timeline identity")
	}
	if !contains(t, out, "Ctrl+X") {
		t.Fatal("running view should show cancel hint in footer")
	}
}

func TestRenderReady(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	out := m.renderReady(Region{Width: 100, Height: 26})
	if !contains(t, out, "OBSERVE") {
		t.Fatal("Ready should show identity")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("Ready should show URL")
	}
	if !contains(t, out, "CC 10") {
		t.Fatal("Ready should show concurrency")
	}
	if !contains(t, out, "Payload") {
		t.Fatal("Ready should show payload state")
	}
	if !contains(t, out, "—") {
		t.Fatal("Ready should show payload as empty (—)")
	}
}

func TestRenderReady_HidesAfterFirstRun(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)

	out := m.View()
	if !contains(t, out, "OBSERVE") {
		t.Fatal("first launch should show Ready surface")
	}

	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out = m.View()
	if contains(t, out, "▶  Ready") {
		t.Fatal("after results exist, Ready should not appear")
	}
}

func TestRenderTopBar_ShowsMethodAndURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderTopBar(m.ShellState(), 100)
	if !contains(t, out, "GET") {
		t.Fatal("top bar should contain method")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("top bar should contain URL")
	}
}

func TestRenderTopBar_ShowsPayloadSummary(t *testing.T) {
	m := NewModel()
	m.shell.Resize(120, 30)
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")
	m.bodyInput.SetValue(`{"key": "value"}`)

	out := m.renderTopBar(m.ShellState(), 120)
	if !contains(t, out, "Payload") {
		t.Fatal("top bar should show payload summary when width permits")
	}
	if !contains(t, out, "1H+B") {
		t.Fatal("top bar should show payload summary with header count and body indicator")
	}
}

func TestRenderTopBar_ShowsCC(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderTopBar(m.ShellState(), 100)
	if !contains(t, out, "CC") {
		t.Fatal("top bar should show CC")
	}
}

func TestRenderTopBar_QueryTruncation(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.urlInput.SetValue("https://api.example.com/users?page=1")
	out := m.renderTopBar(m.ShellState(), 100)
	if contains(t, out, "?page=1") {
		t.Fatal("top bar should truncate query string from URL")
	}
	if !contains(t, out, "api.example.com/users") {
		t.Fatal("top bar should show truncated URL without query")
	}
}

func TestMetricsString_AllStates(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)

	t.Run("idle with no results is empty", func(t *testing.T) {
		m.running = false
		m.summary.Total = 0
		m.summary.SuccessRate = 0
		m.elapsed = 0
		if got := m.metricsString(); got != "" {
			t.Fatalf("idle metrics = %q, want empty", got)
		}
	})

	t.Run("running with no results shows r/s", func(t *testing.T) {
		m.running = true
		m.elapsed = 100 * time.Millisecond
		m.summary.Total = 0
		m.summary.SuccessRate = 0
		out := m.metricsString()
		if out == "" {
			t.Fatal("metrics should appear when running, even with zero results")
		}
		if !contains(t, out, "% ok") {
			t.Fatal("should show success rate")
		}
		if !contains(t, out, "r/s") {
			t.Fatal("should show r/s")
		}
	})

	t.Run("running with results", func(t *testing.T) {
		m.running = true
		m.summary.Total = 50
		m.summary.Successes = 45
		m.summary.SuccessRate = 90
		m.summary.P90 = 100 * time.Millisecond
		m.summary.P99 = 500 * time.Millisecond
		m.summary.MaxLatency = 500 * time.Millisecond
		m.elapsed = 5 * time.Second
		out := m.metricsString()
		for _, want := range []string{"90% ok", "r/s", "p90", "p99"} {
			if !contains(t, out, want) {
				t.Fatalf("metrics should contain %q", want)
			}
		}
	})

	t.Run("100% success rate", func(t *testing.T) {
		m.running = true
		m.summary.Total = 10
		m.summary.Successes = 10
		m.summary.SuccessRate = 100
		m.elapsed = 5 * time.Second
		out := m.metricsString()
		if !contains(t, out, "100% ok") {
			t.Fatal("should show '100% ok'")
		}
	})
}

func TestRenderStatusline_Normal(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[↑↓]") {
		t.Fatal("post-run ribbon should show scroll hint key")
	}
	if !contains(t, out, "[Tab]") {
		t.Fatal("post-run ribbon should show view switch key")
	}
}

func TestRenderStatusline_Ready(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[e] Request") {
		t.Fatal("ready ribbon should show [e] Request")
	}
	if contains(t, out, "[↑↓]") {
		t.Fatal("ready ribbon should not advertise [↑↓] (inert)")
	}
	if contains(t, out, "[Enter]") {
		t.Fatal("ready ribbon should not advertise [Enter] (inert)")
	}
	if contains(t, out, "[Tab]") {
		t.Fatal("ready ribbon should not advertise [Tab] (inert)")
	}
	if !contains(t, out, "[Ctrl+R]") {
		t.Fatal("ready ribbon should show [Ctrl+R]")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("ready ribbon should show [q] Quit")
	}
}

func TestRenderStatusline_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = true
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "Ctrl+X") {
		t.Fatal("running empty ribbon should show Ctrl+X")
	}
	if contains(t, out, "↑↓") {
		t.Fatal("running empty should not advertise ↑↓ (inert)")
	}
	if contains(t, out, "Enter") {
		t.Fatal("running empty should not advertise Enter (inert)")
	}
}

func TestRenderStatusline_RunningWithResults(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = true
	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Enter] Inspect") {
		t.Fatal("running ribbon should show [Enter] Inspect")
	}
	if !contains(t, out, "[Tab] Views") {
		t.Fatal("running ribbon should show [Tab] Views")
	}
	if !contains(t, out, "[Ctrl+X]") {
		t.Fatal("running ribbon should show [Ctrl+X] Cancel")
	}
}

func TestRenderStatusline_RequestDialog(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("request ribbon should show [Esc] Back")
	}
	if !contains(t, out, "[Tab] Next Field") {
		t.Fatal("request ribbon should show [Tab] Next Field")
	}
	if !contains(t, out, "[←→] Method") {
		t.Fatal("request ribbon should show [←→] Method")
	}
}

func TestRenderStatusline_RequestExecDialog(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("request exec ribbon should show [Esc] Back")
	}
	if !contains(t, out, "[↑↓] Adjust") {
		t.Fatal("request exec ribbon should show [↑↓] Adjust")
	}
}

func TestRenderStatusline_Inspecting(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.mode = modeInspect
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("inspect ribbon should show [Esc] Back")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("inspect ribbon should show [q] Quit")
	}
}

func TestRenderStatusline_QuitConfirm(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogConfirmQuit
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "Quit") {
		t.Fatal("ribbon should show quit confirmation")
	}
	if !contains(t, out, "Enter") {
		t.Fatal("quit confirm should mention Enter as confirm option")
	}
}

func TestRenderTimeline_Identity(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderTimeline(Region{Width: 94, Height: 20})
	if !contains(t, out, "Timeline") {
		t.Fatal("timeline should show identity header")
	}
}

func TestRenderTimeline_Empty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderTimeline(Region{Width: 94, Height: 20})
	if !contains(t, out, "Timeline") {
		t.Fatal("empty timeline should show Timeline identity")
	}
	if !contains(t, out, "▶  Ready") {
		t.Fatal("empty timeline should show '▶  Ctrl+R to run'")
	}
}

func TestRenderTimeline_Rows(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
		{Status: 404, Latency: 50 * time.Millisecond},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderTimeline(Region{Width: 94, Height: 20})
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
	m.shell.Resize(100, 24)
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/api"},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderLogs(Region{Width: 94, Height: 20})
	if !contains(t, out, "Logs") {
		t.Fatal("logs should show identity header")
	}
}

func TestRenderLogs_Empty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderLogs(Region{Width: 94, Height: 20})
	if !contains(t, out, "Logs") {
		t.Fatal("empty logs should show Logs identity")
	}
	if !contains(t, out, "▶  Ready") {
		t.Fatal("empty logs should show '▶  Ctrl+R to run'")
	}
}

func TestRenderLogs_Rows(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.urlInput.SetValue("https://example.com/api")
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond, RequestMethod: "GET", RequestURL: "https://example.com/api"},
		{Status: 500, Latency: 50 * time.Millisecond, RequestMethod: "POST", RequestURL: "https://example.com/error"},
	}
	m.summary.MaxLatency = 100 * time.Millisecond
	m.selected = 0

	out := m.renderLogs(Region{Width: 94, Height: 20})
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
	m.shell.Resize(100, 24)
	m.selected = 0
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}

	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "Inspector") {
		t.Fatal("inspector should show identity header")
	}
	if !contains(t, out, "Result #1") {
		t.Fatal("inspector should show result number")
	}
}

func TestRenderInspect_NoSelection(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "No result selected.") {
		t.Fatal("inspector with no selection should show 'No result selected.'")
	}
}

func TestRenderInspect_WithResult(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
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

	out := m.renderInspect(Region{Width: 40, Height: 20})
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
	m.shell.Resize(100, 24)
	m.running = true
	m.selected = 0
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.summary.Total = 50
	m.summary.SuccessRate = 90
	m.elapsed = 5 * time.Second

	out := m.renderInspect(Region{Width: 40, Height: 20})
	if contains(t, out, "% ok") {
		t.Fatal("inspector should NOT show aggregate metrics")
	}
	if contains(t, out, "r/s") {
		t.Fatal("inspector should NOT show requests per second")
	}
}

func TestRenderInspect_WithError(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  500,
			Latency: 100 * time.Millisecond,
			Error:   "connection refused",
		},
	}

	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "connection refused") {
		t.Fatal("inspector should show error message")
	}
	if !contains(t, out, "Error:") {
		t.Fatal("inspector should show 'Error:' prefix")
	}
}

func TestRenderInspect_NoHeaders(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
		},
	}

	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "No headers captured.") {
		t.Fatal("inspector should show 'No headers captured.' when no response headers")
	}
}

func TestRenderInspect_NoBody(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
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

	out := m.renderInspect(Region{Width: 40, Height: 20})
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
	m.shell.Resize(100, 24)
	m.running = true
	m.status = "RUNNING"

	out := m.renderTimeline(Region{Width: 94, Height: 20})
	if !contains(t, out, "Timeline") {
		t.Fatal("running empty timeline should show Timeline identity")
	}
	if !contains(t, out, "⏳  Waiting for results...") {
		t.Fatal("running empty timeline should show '⏳  Waiting for results...'")
	}
}

func TestRenderTimeline_IdleNoURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = false
	m.urlInput.SetValue("")

	out := m.renderTimeline(Region{Width: 94, Height: 20})
	if !contains(t, out, "Timeline") {
		t.Fatal("idle empty timeline should show Timeline identity")
	}
	if !contains(t, out, "Enter a URL to begin") {
		t.Fatal("idle empty timeline with no URL should show 'Enter a URL to begin'")
	}
}

func TestRenderTimeline_IdleWithURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = false
	m.urlInput.SetValue("https://example.com/api")

	out := m.renderTimeline(Region{Width: 94, Height: 20})
	if !contains(t, out, "Timeline") {
		t.Fatal("idle empty timeline should show Timeline identity")
	}
	if !contains(t, out, "▶  Ready") {
		t.Fatal("idle empty timeline with URL should show '▶  Ctrl+R to run'")
	}
}

func TestRenderLogs_RunningEmpty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = true
	m.status = "RUNNING"

	out := m.renderLogs(Region{Width: 94, Height: 20})
	if !contains(t, out, "Logs") {
		t.Fatal("running empty logs should show Logs identity")
	}
	if !contains(t, out, "📭  No results yet...") {
		t.Fatal("running empty logs should show '📭  No results yet...'")
	}
}

func TestRenderLogs_IdleNoURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = false
	m.urlInput.SetValue("")

	out := m.renderLogs(Region{Width: 94, Height: 20})
	if !contains(t, out, "Logs") {
		t.Fatal("idle empty logs should show Logs identity")
	}
	if !contains(t, out, "Enter a URL to begin") {
		t.Fatal("idle empty logs with no URL should show 'Enter a URL to begin'")
	}
}

func TestRenderLogs_IdleWithURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = false
	m.urlInput.SetValue("https://example.com/api")

	out := m.renderLogs(Region{Width: 94, Height: 20})
	if !contains(t, out, "Logs") {
		t.Fatal("idle empty logs should show Logs identity")
	}
	if !contains(t, out, "▶  Ready") {
		t.Fatal("idle empty logs with URL should show '▶  Ctrl+R to run'")
	}
}

func TestView_WidthMinClamp(t *testing.T) {
	m := NewModel()
	m.shell.Resize(40, 30)

	out := m.View()
	if !contains(t, out, "GET") {
		t.Fatal("View should contain method even at small width")
	}
}

func TestView_PayloadNotShown(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)

	out := m.View()
	if contains(t, out, "HEADERS") {
		t.Fatal("View should not contain HEADERS when payload dialog is closed")
	}
	if contains(t, out, "BODY") {
		t.Fatal("View should not contain BODY when payload dialog is closed")
	}
}

func TestRenderRequest_ShowsBadgeHeadersBody(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("request should show identity header")
	}
	if !contains(t, out, "HEADERS") {
		t.Fatal("request should contain 'HEADERS'")
	}
	if !contains(t, out, "BODY") {
		t.Fatal("request should contain 'BODY'")
	}
	if !contains(t, out, "Content-Type") {
		t.Fatal("request should contain header key")
	}
	if !contains(t, out, "application/json") {
		t.Fatal("request should contain header value")
	}
}

func TestRenderRequest_ShowsNoHeadersMessage(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("request should show identity header")
	}
	if !contains(t, out, "No headers configured.") {
		t.Fatal("request should show 'No headers configured.' when no headers")
	}
}

func TestRenderRequest_ShowsBodyPlaceholder(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("payload should show identity header")
	}
	if !contains(t, out, `{"name":"pulse"}`) {
		t.Fatal("payload should show body placeholder when body is empty")
	}
}

func TestRenderRequest_Identity(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()
	out := m.renderRequest(Region{Width: 100, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("request should show identity header")
	}
	if !contains(t, out, "URL") {
		t.Fatal("request should show URL label")
	}
	if !contains(t, out, "Method") {
		t.Fatal("request should show Method label")
	}
	if !contains(t, out, "GET") {
		t.Fatal("request should show method options")
	}
}

func TestRenderRequest_ExecutionIdentity(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.setConcurrency(7)
	out := m.renderRequest(Region{Width: 100, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("request should show identity header")
	}
	if !contains(t, out, "7") {
		t.Fatal("execution should show current concurrency value")
	}
	if !contains(t, out, "1–100") {
		t.Fatal("execution should show range affordance")
	}
}

func TestRenderRequest_RequestDomain_Focused(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()
	m.urlInput.SetValue("https://example.com/api")

	out := m.renderRequest(Region{Width: 100, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("should show identity")
	}
	if !contains(t, out, "api") {
		t.Fatal("should show URL")
	}
	if !m.urlInput.Focused() {
		t.Fatal("urlInput should be focused when dialog is open")
	}
}

func TestRenderRequest_ExecDomain_Focused(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.ccInput.Focus()
	m.setConcurrency(7)

	out := m.renderRequest(Region{Width: 100, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("should show identity")
	}
	if !contains(t, out, "1–100") {
		t.Fatal("should show range")
	}
	if !m.ccInput.Focused() {
		t.Fatal("ccInput should be focused when dialog is open")
	}
}

func TestRenderRequest_PayloadDomain_HeaderKeyFocus(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Key.Focus()
	m.headers[0].Value.SetValue("application/json")

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "Content-Type") {
		t.Fatal("should show header key")
	}
	if !contains(t, out, "application/json") {
		t.Fatal("should show header value")
	}
	if !m.headers[0].Key.Focused() {
		t.Fatal("key should be focused when subfocus is subfocusKey")
	}
	if m.headers[0].Value.Focused() {
		t.Fatal("value should NOT be focused when subfocus is subfocusKey")
	}
}

func TestRenderRequest_PayloadDomain_HeaderValueFocus(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusValue
	m.headers = append(m.headers, newHeaderRow())
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")
	m.headers[0].Value.Focus()

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "Content-Type") {
		t.Fatal("should show header key")
	}
	if !contains(t, out, "application/json") {
		t.Fatal("should show header value")
	}
	if m.headers[0].Key.Focused() {
		t.Fatal("key should NOT be focused when subfocus is subfocusValue")
	}
	if !m.headers[0].Value.Focused() {
		t.Fatal("value should be focused when subfocus is subfocusValue")
	}
}

func TestRenderRequest_PayloadDomain_BodyFocus(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.bodyInput.Focus()
	m.bodyInput.SetValue(`{"key": "value"}`)
	m.headers = append(m.headers, newHeaderRow())

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, `{"key": "value"}`) {
		t.Fatal("should show body content")
	}
	if !m.bodyInput.Focused() {
		t.Fatal("bodyInput should be focused when selectedHead is bodyFocus")
	}
}

func TestRenderCurrentSurface_DispatchesToSurface(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)

	region := Region{Width: 100, Height: 26}

	// Ready state
	out := m.resolveSurface().Render(region)
	if !contains(t, out, "OBSERVE") {
		t.Fatal("renderCurrentSurface should render Ready when idle")
	}
}

func TestRenderWorkspace_InspectorDrillDown(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.mode = modeInspect
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
	m.shell.Resize(100, 24)
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

func TestRenderRequest_PayloadDomain_SelectedRowVisible(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow(), newHeaderRow())
	m.headers[0].Key.SetValue("Authorization")
	m.headers[0].Value.SetValue("Bearer token")
	m.headers[1].Key.SetValue("Content-Type")
	m.headers[1].Value.SetValue("application/json")
	m.selectedHead = 1

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "Authorization") {
		t.Fatal("request should show first header key")
	}
	if !contains(t, out, "Content-Type") {
		t.Fatal("request should show second header key")
	}
	// selected row should have cursor
	if !contains(t, out, "▶ Content-Type") {
		t.Fatal("selected header row should show ▶ cursor")
	}
}

func TestRenderRequest_PayloadDomain_BodyFocusColor(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.headers = append(m.headers, newHeaderRow())

	out := m.renderRequest(Region{Width: 96, Height: 20})
	if !contains(t, out, "BODY") {
		t.Fatal("request should show BODY label")
	}
}

func TestRenderStatusline_PayloadDomain(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Tab]") {
		t.Fatal("request payload ribbon should show [Tab]")
	}
	if !contains(t, out, "[Ctrl+N]") {
		t.Fatal("request payload ribbon should show [Ctrl+N]")
	}
	if !contains(t, out, "[Ctrl+D]") {
		t.Fatal("request payload ribbon should show [Ctrl+D]")
	}
	if !contains(t, out, "[Esc]") {
		t.Fatal("request payload ribbon should show [Esc]")
	}
}

func TestConfirmQuit_PreservesWorkspace(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.running = true
	m.workspace.dialog = dialogConfirmQuit

	out := m.View()
	if !contains(t, out, "Ctrl+C") {
		t.Fatal("confirm quit should show ctrl+c prompt")
	}
	// Body content should still be visible (Timeline identity preserved)
	if !contains(t, out, "Timeline") {
		t.Fatal("confirm quit should preserve workspace identity")
	}
}

func TestPayloadSummary(t *testing.T) {
	m := NewModel()
	tt := []struct {
		headers int
		body    string
		want    string
	}{
		{0, "", "—"},
		{1, "", "1H"},
		{0, "body", "B"},
		{2, "body", "2H+B"},
	}
	for _, tc := range tt {
		m.headers = nil
		for i := 0; i < tc.headers; i++ {
			m.headers = append(m.headers, newHeaderRow())
		}
		m.bodyInput.SetValue(tc.body)
		got := m.payloadSummary()
		if got != tc.want {
			t.Errorf("payloadSummary(%d headers, %q) = %q (expected %q)", tc.headers, tc.body, got, tc.want)
		}
	}
}

func TestOrientationLabel_AllStates(t *testing.T) {
	tt := []struct {
		name   string
		setup  func(m *Model)
		expect string
	}{
		{"Ready", func(m *Model) {}, "OBSERVE"},
		{"WithResults", func(m *Model) { m.results = []model.Result{{Status: 200}} }, "OBSERVE"},
		{"RunningEmpty", func(m *Model) { m.running = true }, "OBSERVE"},
		{"RunningWithResults", func(m *Model) { m.running = true; m.results = []model.Result{{Status: 200}} }, "OBSERVE"},
		{"LogsView", func(m *Model) { m.workspace.view = LogsView; m.results = []model.Result{{Status: 200}} }, "OBSERVE"},
		{"RunningLogsView", func(m *Model) {
			m.running = true
			m.workspace.view = LogsView
			m.results = []model.Result{{Status: 200}}
		}, "OBSERVE"},
		{"RequestDialog", func(m *Model) { m.workspace.dialog = dialogRequest }, "REQUEST"},
		{"ExecDomain", func(m *Model) { m.workspace.dialog = dialogRequest; m.activeDomain = DomainExec }, "REQUEST"},
		{"PayloadDomain", func(m *Model) { m.workspace.dialog = dialogRequest; m.activeDomain = DomainPayload }, "REQUEST"},
		{"InspectMode", func(m *Model) { m.workspace.mode = modeInspect }, "INSPECT"},
		{"QuitDialog", func(m *Model) { m.workspace.dialog = dialogConfirmQuit }, "QUIT"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(100, 24)
			tc.setup(&m)
			if got := orientationLabel(m); got != tc.expect {
				t.Fatalf("orientationLabel = %q, want %q", got, tc.expect)
			}
		})
	}
}

func TestRenderStatusline_EmptyGroupsOmitted(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	// Ready state has no Navigation commands — must not render "[↑↓]" or "[Enter] Inspect".
	if contains(t, out, "[↑↓]") {
		t.Fatal("ribbon should omit empty Navigation group in ready state")
	}
}

func TestRenderStatusline_CategoryOrder(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out := m.renderStatusline(m.ShellState(), 100)
	// With results: Navigation → Configuration → Operation → Application.
	// Navigation must appear before Configuration.
	navIdx := strings.Index(out, "[↑↓]")
	cfgIdx := strings.Index(out, "[e]")
	if navIdx < 0 {
		t.Fatal("ribbon should include [↑↓] when results exist")
	}
	if cfgIdx < 0 {
		t.Fatal("ribbon should include [e] when results exist")
	}
	if navIdx > cfgIdx {
		t.Fatal("Navigation group must render before Configuration group")
	}
}

func TestRenderStatusline_WithinGroupSeparator(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	// Ready state: [e] Request in Configuration group.
	if !contains(t, out, "[e] Request") {
		t.Fatal("ready state should show [e] Request")
	}
}

func TestRenderStatusline_BetweenGroupSeparator(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	// Ready state: Configuration group followed by Operation group.
	// Must have 4-space gap between groups.
	if !contains(t, out, "Request    [Ctrl+R]") {
		t.Fatal("different category groups must be separated by wider gap (4 spaces)")
	}
}

// ---------------------------------------------------------------------------
// Architectural invariant tests — ownership rules, not content
// ---------------------------------------------------------------------------

// hasFullRule reports whether any line in s consists entirely of ─ characters,
// identifying Shell-owned full-width separators (not typographic domain headers).
func hasFullRule(s string) bool {
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.TrimRight(trimmed, "─") == "" {
			return true
		}
	}
	return false
}

// TestShellInvariant_WorkspaceNoSeparators verifies workspace surface renderers
// never produce full-width shell separators (lines of only ─). Typographic
// domain headers (── Payload ──) are allowed because they are workspace-owned.
func TestShellInvariant_WorkspaceNoSeparators(t *testing.T) {
	region := Region{Width: 100, Height: 26}

	// renderReady
	m := NewModel()
	m.shell.Resize(100, 24)
	if hasFullRule(m.renderReady(region)) {
		t.Fatal("renderReady must not render full shell separators")
	}

	// renderRequest
	m2 := NewModel()
	m2.shell.Resize(100, 24)
	m2.workspace.dialog = dialogRequest
	m2.activeDomain = DomainRequest
	if hasFullRule(m2.renderRequest(region)) {
		t.Fatal("renderRequest must not render full shell separators")
	}

	// renderTimeline
	m3 := NewModel()
	m3.shell.Resize(100, 24)
	m3.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	if hasFullRule(m3.renderTimeline(region)) {
		t.Fatal("renderTimeline must not render full shell separators")
	}

	// renderLogs
	m4 := NewModel()
	m4.shell.Resize(100, 24)
	m4.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	if hasFullRule(m4.renderLogs(region)) {
		t.Fatal("renderLogs must not render full shell separators")
	}

	// renderInspect
	m5 := NewModel()
	m5.shell.Resize(100, 24)
	m5.selected = 0
	m5.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	if hasFullRule(m5.renderInspect(Region{Width: 40, Height: 20})) {
		t.Fatal("renderInspect must not render full shell separators")
	}
}

// TestShellInvariant_WorkspaceNoShortcuts verifies workspace surface renderers
// never bake keyboard shortcuts into body content. Shortcuts belong in the
// ribbon, which is Shell-owned.
func TestShellInvariant_WorkspaceNoShortcuts(t *testing.T) {
	region := Region{Width: 100, Height: 26}

	shortcutPatterns := []string{"Ctrl+", "[Esc]", "[Tab]", "[Enter]", "[↑↓]", "[←→]", "[q]", "[e]"}

	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.resolveSurface().Render(region)
	for _, pat := range shortcutPatterns {
		if contains(t, out, pat) {
			t.Fatalf("Workspace must not contain %q (shortcuts belong in ribbon)", pat)
		}
	}

	// Verify the REQUEST surface also follows this rule.
	m2 := NewModel()
	m2.shell.Resize(100, 24)
	m2.workspace.dialog = dialogRequest
	m2.activeDomain = DomainRequest
	for _, pat := range shortcutPatterns {
		if contains(t, m2.renderRequest(region), pat) {
			t.Fatalf("Request surface must not contain %q", pat)
		}
	}
}

// TestShellInvariant_RibbonHasOrientation verifies every ribbon output starts
// with a known orientation label (the Shell Column).
func TestShellInvariant_RibbonHasOrientation(t *testing.T) {
	labels := []string{"OBSERVE", "REQUEST", "INSPECT", "QUIT"}
	hasLabel := func(out string) bool {
		for _, l := range labels {
			if contains(t, out, l) {
				return true
			}
		}
		return false
	}

	m := NewModel()
	m.shell.Resize(100, 24)

	// Idle
	if !hasLabel(m.renderStatusline(m.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label in idle state")
	}

	// Running empty
	m2 := NewModel()
	m2.shell.Resize(100, 24)
	m2.running = true
	if !hasLabel(m2.renderStatusline(m2.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label when running empty")
	}

	// With results
	m3 := NewModel()
	m3.shell.Resize(100, 24)
	m3.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	if !hasLabel(m3.renderStatusline(m3.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label with results")
	}

	// Inspect mode
	m4 := NewModel()
	m4.shell.Resize(100, 24)
	m4.workspace.mode = modeInspect
	if !hasLabel(m4.renderStatusline(m4.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label in inspect mode")
	}

	// Request dialog
	m5 := NewModel()
	m5.shell.Resize(100, 24)
	m5.workspace.dialog = dialogRequest
	if !hasLabel(m5.renderStatusline(m5.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label in request dialog")
	}

	// ConfirmQuit dialog
	m6 := NewModel()
	m6.shell.Resize(100, 24)
	m6.workspace.dialog = dialogConfirmQuit
	if !hasLabel(m6.renderStatusline(m6.ShellState(), 100)) {
		t.Fatal("ribbon must show orientation label in quit dialog")
	}
}

// TestShellInvariant_ActionsAreIntents verifies Actions() returns behavioral
// Action values, not presentation strings. Every Action must have a valid
// ActionID that exists in the actionBindings table.
func TestShellInvariant_ActionsAreIntents(t *testing.T) {
	states := []func() []Action{
		func() []Action { m := NewModel(); return m.Actions() },
		func() []Action { m := NewModel(); m.running = true; return m.Actions() },
		func() []Action { m := NewModel(); m.running = true; m.results = []model.Result{{}}; return m.Actions() },
		func() []Action { m := NewModel(); m.results = []model.Result{{}}; return m.Actions() },
		func() []Action {
			m := NewModel()
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainRequest
			return m.Actions()
		},
		func() []Action {
			m := NewModel()
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainExec
			return m.Actions()
		},
		func() []Action {
			m := NewModel()
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainPayload
			return m.Actions()
		},
		func() []Action { m := NewModel(); m.workspace.mode = modeInspect; return m.Actions() },
		func() []Action { m := NewModel(); m.workspace.dialog = dialogConfirmQuit; return m.Actions() },
	}

	for i, actions := range states {
		acts := actions()
		if len(acts) == 0 {
			t.Fatalf("state %d: Actions() must not return empty", i)
		}
		for _, a := range acts {
			if _, ok := actionBindings[a.ID]; !ok {
				t.Fatalf("state %d: Action ID %v has no binding", i, a.ID)
			}
			if !a.Enabled {
				t.Fatalf("state %d: Action %v should be enabled", i, a.ID)
			}
		}
	}
}

// TestShellInvariant_ViewOwnsSeparators verifies the shell (View) renders
// separators, and they appear exactly twice (context separator, ribbon separator).
func TestShellInvariant_ViewOwnsSeparators(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	out := m.View()
	if !contains(t, out, "─") {
		t.Fatal("View must render shell separators")
	}
}

// ---------------------------------------------------------------------------
// ShellState snapshot tests
// ---------------------------------------------------------------------------

func TestShellState_ContainsOrientation(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	s := m.ShellState()
	if s.Orientation == "" {
		t.Fatal("ShellState.Orientation must not be empty")
	}
}

func TestShellState_ContainsConfiguration(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	s := m.ShellState()
	if len(s.Configuration) == 0 {
		t.Fatal("ShellState.Configuration must not be empty")
	}
	foundMethod := false
	for _, c := range s.Configuration {
		if c.Identity == "Method" {
			foundMethod = true
		}
	}
	if !foundMethod {
		t.Fatal("ShellState.Configuration must contain Method")
	}
}

func TestShellState_ContainsActions(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	s := m.ShellState()
	if len(s.Actions) == 0 {
		t.Fatal("ShellState.Actions must not be empty")
	}
}

func TestShellState_QueryTruncated(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.urlInput.SetValue("https://api.example.com/users?page=1")
	s := m.ShellState()
	url := ""
	for _, c := range s.Configuration {
		if c.Identity == "URL" {
			url = c.Value
		}
	}
	if !contains(t, url, "?page=1") {
		t.Fatal("ShellState URL should retain full value (truncation is a renderer concern)")
	}
}

func TestShellState_UpdatesWithState(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.running = true
	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}

	s := m.ShellState()
	if s.Orientation != "OBSERVE" {
		t.Fatalf("running state ShellState.Orientation = %q, want OBSERVE", s.Orientation)
	}
}

// ---------------------------------------------------------------------------
// Configuration model tests
// ---------------------------------------------------------------------------

func TestConfiguration_MethodAndURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	cfg := m.Configuration()
	if len(cfg) < 3 {
		t.Fatal("Configuration should have at least 3 items (Method, URL, CC)")
	}
	if cfg[0].Identity != "Method" || cfg[0].Value == "" {
		t.Fatal("Configuration[0] should be Method with a value")
	}
	if cfg[1].Identity != "URL" || cfg[1].Value == "" {
		t.Fatal("Configuration[1] should be URL with a value")
	}
	if cfg[2].Identity != "CC" || cfg[2].Value == "" {
		t.Fatal("Configuration[2] should be CC with a value")
	}
}

func TestConfiguration_InvalidURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.urlInput.SetValue("")
	cfg := m.Configuration()
	for _, c := range cfg {
		if c.Identity == "URL" && c.Valid {
			t.Fatal("empty URL should be marked invalid")
		}
	}
}

func TestConfiguration_PayloadIncluded(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	m.headers = append(m.headers, newHeaderRow())
	m.bodyInput.SetValue("{}")
	cfg := m.Configuration()
	found := false
	for _, c := range cfg {
		if c.Identity == "Payload" {
			found = true
		}
	}
	if !found {
		t.Fatal("Configuration should include Payload when headers or body exist")
	}
}

func TestConfiguration_PayloadExcluded(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	cfg := m.Configuration()
	for _, c := range cfg {
		if c.Identity == "Payload" {
			t.Fatal("Configuration should NOT include Payload when empty")
		}
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

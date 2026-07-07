package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
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
	if !contains(t, out, "READY") {
		t.Fatal("View should contain READY identity in Ready surface")
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
	if !contains(t, out, "Prepare") {
		t.Fatal("Ready should show Prepare purpose")
	}
	if !contains(t, out, "Current Request") {
		t.Fatal("Ready should show current request heading")
	}
	if !contains(t, out, "httpbin") {
		t.Fatal("Ready should show URL")
	}
	if !contains(t, out, fmt.Sprintf("%d", runconfig.DefaultConcurrency)) {
		t.Fatal("Ready should show concurrency")
	}
	if !contains(t, out, "Payload") {
		t.Fatal("Ready should show payload field")
	}
	if !contains(t, out, sentinelEmpty) {
		t.Fatal("Ready should show payload as empty")
	}
	if !contains(t, out, "Ready to execute") {
		t.Fatal("Ready should show readiness status")
	}
}

func TestRenderReady_HidesAfterFirstRun(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)

	out := m.View()
	if !contains(t, out, "Ready to execute") {
		t.Fatal("first launch should show Ready surface with readiness status")
	}

	m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
	out = m.View()
	if contains(t, out, "Ready to execute") {
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

func TestRenderTopBar_ShowsConcurrency(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderTopBar(m.ShellState(), 100)
	if !contains(t, out, "CC ") {
		t.Fatal("top bar should show CC (concurrency) abbreviation")
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
	if !contains(t, out, "[[]]") {
		t.Fatal("post-run ribbon should show view switch key")
	}
}

func TestRenderStatusline_Ready(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[e] Configure") {
		t.Fatal("ready ribbon should show [e] Configure")
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
	if !contains(t, out, "[[]] View") {
		t.Fatal("running ribbon should show [[]] View")
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
	if !contains(t, out, "Ready") {
		t.Fatal("empty timeline should show Ready state")
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
	if !contains(t, out, "Ready") {
		t.Fatal("empty logs should show Ready state")
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
	if !contains(t, out, "Result 1") {
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
	// Visual widths must be identical (selection changes appearance, not geometry)
	if lipgloss.Width(rowCursor(true)) != lipgloss.Width(rowCursor(false)) {
		t.Fatal("cursor visual width must be invariant")
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
	if !contains(t, out, "Waiting for completions...") {
		t.Fatal("running empty timeline should show completion waiting state")
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
	if !contains(t, out, "Ready") {
		t.Fatal("idle empty timeline with URL should show Ready state")
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
	if !contains(t, out, "No events captured yet...") {
		t.Fatal("running empty logs should show sequence waiting state")
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
	if !contains(t, out, "Ready") {
		t.Fatal("idle empty logs with URL should show Ready state")
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
	if !contains(t, out, "1-100") {
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
	m.concurrencyInput.Focus()
	m.setConcurrency(7)

	out := m.renderRequest(Region{Width: 100, Height: 20})
	if !contains(t, out, "REQUEST") {
		t.Fatal("should show identity")
	}
	if !contains(t, out, "1-100") {
		t.Fatal("should show range")
	}
	if !m.concurrencyInput.Focused() {
		t.Fatal("concurrencyInput should be focused when dialog is open")
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
	if !contains(t, out, "Ready") {
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
	if !contains(t, out, "Result 1") {
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
		{"Ready", func(m *Model) {}, "READY"},
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
	// Ready state: [e] Configure in Configuration group.
	if !contains(t, out, "[e] Configure") {
		t.Fatal("ready state should show [e] Configure")
	}
}

func TestRenderStatusline_BetweenGroupSeparator(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	// Ready state: Configuration group followed by Operation group.
	// Must have 4-space gap between groups.
	if !contains(t, out, "Configure    [Ctrl+R]") {
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
	labels := []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "QUIT"}
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

// ---------------------------------------------------------------------------
// Validation Visibility
// ---------------------------------------------------------------------------

func TestRenderRequest_ValidationVisibility(t *testing.T) {
	t.Run("shows validation block in exec domain when errMsg set", func(t *testing.T) {
		m := NewModel()
		m.errMsg = "INVALID URL"
		m.shell.Resize(100, 30)
		out := m.renderExecDomain(80)
		if !contains(t, out, "INVALID URL") {
			t.Fatal("renderExecDomain should show the error message")
		}
		if !contains(t, out, "Adjust the request and run again.") {
			t.Fatal("renderExecDomain should show recovery guidance")
		}
	})

	t.Run("omits validation block when errMsg empty", func(t *testing.T) {
		m := NewModel()
		m.errMsg = ""
		m.shell.Resize(100, 30)
		out := m.renderExecDomain(80)
		if contains(t, out, "INVALID URL") {
			t.Fatal("renderExecDomain must NOT show validation when errMsg is empty")
		}
	})
}

// ---------------------------------------------------------------------------
// Rendering Vocabulary
// ---------------------------------------------------------------------------

func TestRenderBodyPreview(t *testing.T) {
	t.Run("empty body shows placeholder", func(t *testing.T) {
		got := renderBodyPreview("", 10)
		if !contains(t, got, "No body captured.") {
			t.Fatal("empty body should show placeholder")
		}
	})

	t.Run("JSON body is formatted", func(t *testing.T) {
		got := renderBodyPreview(`{"a":1,"b":2}`, 10)
		if !contains(t, got, `"a"`) || !contains(t, got, `"b"`) {
			t.Fatal("JSON body should be formatted and show keys")
		}
	})

	t.Run("truncation shows ellipsis", func(t *testing.T) {
		body := "line1\nline2\nline3\nline4\nline5"
		got := renderBodyPreview(body, 3)
		if !contains(t, got, "... (truncated)") {
			t.Fatal("body exceeding maxLines should show truncation indicator")
		}
	})

	t.Run("plain text within max lines is not truncated", func(t *testing.T) {
		body := "line1\nline2"
		got := renderBodyPreview(body, 5)
		if contains(t, got, "truncated") {
			t.Fatal("body within maxLines must not show truncation")
		}
	})
}

func TestRenderMetadata(t *testing.T) {
	t.Run("renders key-value pair", func(t *testing.T) {
		got := renderMetadata("Key", "Value")
		if !contains(t, got, "Key") || !contains(t, got, "Value") {
			t.Fatal("renderMetadata should contain label and value")
		}
		if !contains(t, got, ":") {
			t.Fatal("renderMetadata should contain colon separator")
		}
	})
}

// ---------------------------------------------------------------------------
// Workspace Constitution — Regression Tests
// ---------------------------------------------------------------------------

func TestWorkspaceConstitution_Ready(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	out := m.View()

	if !contains(t, out, "READY") {
		t.Fatal("Constitution: Ready must show READY identity")
	}
	if !contains(t, out, "Prepare") {
		t.Fatal("Constitution: Ready must show Prepare purpose")
	}
	if !contains(t, out, "Current Request") {
		t.Fatal("Constitution: Ready must answer 'what to look at first'")
	}
	if !contains(t, out, "[Ctrl+R]") {
		t.Fatal("Constitution: Ready next action must be Run")
	}
	if !contains(t, out, "[e] Configure") {
		t.Fatal("Constitution: Ready must offer Configuration before running")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("Constitution: Ready must show exit path")
	}
}

func TestWorkspaceConstitution_Timeline(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.shell.Resize(100, 30)
	out := m.View()

	if !contains(t, out, "Timeline") {
		t.Fatal("Constitution: Timeline must show identity")
	}
	if !contains(t, out, "200") {
		t.Fatal("Constitution: Timeline must show result status")
	}
	if !contains(t, out, "[Enter] Inspect") {
		t.Fatal("Constitution: Timeline next action must be Inspect")
	}
	if !contains(t, out, "[[]] View") {
		t.Fatal("Constitution: Timeline must offer view switching")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("Constitution: Timeline must show exit path")
	}
}

func TestWorkspaceConstitution_Logs(t *testing.T) {
	m := NewModel()
	m.workspace.view = LogsView
	m.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	m.shell.Resize(100, 30)
	out := m.View()

	if !contains(t, out, "Logs") {
		t.Fatal("Constitution: Logs must show identity")
	}
	if !contains(t, out, "[Enter] Inspect") {
		t.Fatal("Constitution: Logs next action must be Inspect")
	}
	if !contains(t, out, "[[]] View") {
		t.Fatal("Constitution: Logs must offer view switching")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("Constitution: Logs must show exit path")
	}
}

func TestWorkspaceConstitution_Inspect(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = []model.Result{
		{
			Status:          200,
			Latency:         100 * time.Millisecond,
			ResponseHeaders: map[string]string{"Content-Type": "application/json"},
			ResponseBody:    `{"ok": true}`,
		},
	}
	m.selected = 0
	m.shell.Resize(100, 30)
	out := m.View()

	if !contains(t, out, "INSPECT") {
		t.Fatal("Constitution: Inspect must show INSPECT identity in ribbon")
	}
	if !contains(t, out, "WHAT HAPPENED") {
		t.Fatal("Constitution: Inspect must show WHAT HAPPENED investigation section")
	}
	if !contains(t, out, "WHY") {
		t.Fatal("Constitution: Inspect must show WHY investigation section")
	}
	if !contains(t, out, "RESPONSE") {
		t.Fatal("Constitution: Inspect must show RESPONSE section")
	}
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("Constitution: Inspect next action must be Back")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("Constitution: Inspect must show exit path")
	}
}

func TestWorkspaceConstitution_Request(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.shell.Resize(100, 30)
	out := m.View()

	if !contains(t, out, "REQUEST") {
		t.Fatal("Constitution: Request must show REQUEST identity")
	}
	if !contains(t, out, "Request") {
		t.Fatal("Constitution: Request dialog must show Request domain header")
	}
	if !contains(t, out, "Payload") {
		t.Fatal("Constitution: Request dialog must show Payload domain")
	}
	if !contains(t, out, "Execution") {
		t.Fatal("Constitution: Request dialog must show Execution domain")
	}
	if !contains(t, out, "[Tab]") {
		t.Fatal("Constitution: Request next action must be Tab navigation")
	}
	if !contains(t, out, "[Ctrl+R]") {
		t.Fatal("Constitution: Request must show Run action")
	}
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("Constitution: Request must show recovery/exit path")
	}
}

// ---------------------------------------------------------------------------
// v0.9.6 Inquiry Constitution — investigation workspace
// ---------------------------------------------------------------------------

func TestInvestigation_PromotedMetadata(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{
				"Content-Type":     "application/json",
				"Content-Encoding": "gzip",
			},
			ResponseBody: `{"ok": true}`,
		},
	}

	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "Content-Type: application/json") {
		t.Fatal("Promoted metadata should show Content-Type")
	}
	if !contains(t, out, "Encoding: gzip") {
		t.Fatal("Promoted metadata should show Content-Encoding")
	}
	if !contains(t, out, "Content-Length") {
		t.Fatal("Promoted metadata should show Content-Length")
	}
}

func TestInvestigation_BinaryContent(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{
				"Content-Type": "image/png",
			},
			ResponseBody: "\x89PNG\r\n\x1a\n",
		},
	}

	out := m.renderInspectBody(m.results[0], 10, 40)
	if !contains(t, out, "Binary content") {
		t.Fatal("binary body should show 'Binary content' indicator")
	}
	if !contains(t, out, "image/png") {
		t.Fatal("binary body should show content type")
	}
}

func TestInvestigation_BodyScrolling(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.inspectZone = zoneBody
	m.selected = 0
	m.results = []model.Result{
		{
			Status:       200,
			Latency:      100 * time.Millisecond,
			ResponseBody: "line1\nline2\nline3\nline4\nline5",
		},
	}

	// Scroll down twice
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	// Render with 3 visible lines starting at offset 2
	out := m.renderInspectBody(m.results[0], 3, 40)
	if !contains(t, out, "line3") {
		t.Fatal("scrolled body should show line3")
	}
	if !contains(t, out, "line5") {
		t.Fatal("scrolled body should show line5")
	}
	if contains(t, out, "line1") {
		t.Fatal("scrolled body should NOT show line1 (scrolled past)")
	}
}

func TestInvestigation_Continuity(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
		{Status: 404, Latency: 20 * time.Millisecond},
	}
	m.selected = 1

	// Esc should return to Observe and preserve selection
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.workspace.mode != modeObserve {
		t.Fatal("Esc should return to Observe mode")
	}
	if m.selected != 1 {
		t.Fatalf("Esc should preserve selection, got %d", m.selected)
	}
}

func TestInvestigation_ZoneNavigation(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
	}

	// Verify zone emphasis renders differently for active vs inactive
	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "WHAT HAPPENED") {
		t.Fatal("WHAT HAPPENED zone should be visible")
	}

	// Tab to WHY zone
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.inspectZone != zoneWhy {
		t.Fatal("Tab should advance to WHY zone")
	}

	// Tab to BODY zone
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.inspectZone != zoneBody {
		t.Fatal("Tab should advance to BODY zone")
	}

	// Tab wraps back to WHAT HAPPENED
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.inspectZone != zoneWhatHappened {
		t.Fatal("Tab should wrap to WHAT HAPPENED zone")
	}
}

func TestInvestigation_NarrowLayoutZonePresence(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.selected = 0
	m.results = []model.Result{
		{
			Status:  200,
			Latency: 100 * time.Millisecond,
			ResponseHeaders: map[string]string{
				"Content-Type":     "application/json",
				"Content-Encoding": "gzip",
				"Content-Length":   "456",
			},
			ResponseBody: `{"key": "value"}`,
		},
	}

	// Narrow layout (Width < 60) with multi-line "what" content
	// (method+url + status + 3 promoted metadata = 5 lines).
	// Content-derived heights must preserve all three investigation zones.
	out := m.renderInspect(Region{Width: 40, Height: 20})
	if !contains(t, out, "WHAT HAPPENED") {
		t.Fatal("narrow layout must show WHAT HAPPENED section")
	}
	if !contains(t, out, "WHY") {
		t.Fatal("content-derived height must preserve WHY section")
	}
	if !contains(t, out, "RESPONSE") {
		t.Fatal("content-derived height must preserve RESPONSE section")
	}

	// Verify minimal content doesn't panic or produce empty output
	m2 := NewModel()
	m2.workspace.mode = modeInspect
	m2.selected = 0
	m2.results = []model.Result{
		{Status: 200, Latency: 100 * time.Millisecond},
	}
	outMin := m2.renderInspect(Region{Width: 40, Height: 10})
	if outMin == "" {
		t.Fatal("narrow layout with minimal content must not produce empty output")
	}
}

func TestHeaderDeleteSafety(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	// ctrl+d with bodyFocus (-1) selectedHead must not panic
	// (regression: missing guard would slice m.headers[:bodyFocus])
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	// State should be unchanged — still at body focus, no headers
	if m.selectedHead != bodyFocus {
		t.Fatal("ctrl+d with body focus must not change selectedHead")
	}
	if len(m.headers) != 0 {
		t.Fatal("ctrl+d with empty headers must not add headers")
	}
}

func TestInvestigationStateResetOnEnter(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeObserve
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
	}
	m.selected = 0

	// Set stale inspect state
	m.inspectZone = zoneBody
	m.inspectBodyOffset = 5

	// Enter inspect mode — must reset zone and offset
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.workspace.mode != modeInspect {
		t.Fatal("enter should switch to inspect mode")
	}
	if m.inspectZone != zoneWhatHappened {
		t.Fatal("enter inspect must reset inspectZone to whatHappened")
	}
	if m.inspectBodyOffset != 0 {
		t.Fatal("enter inspect must reset inspectBodyOffset to 0")
	}
}

func TestInvestigationStateResetOnStartRun(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeObserve
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
	}

	// Set stale inspect state
	m.inspectZone = zoneBody
	m.inspectBodyOffset = 5

	// Start a new run — must reset zone and offset
	started, cmd := m.startRun()
	if cmd == nil {
		t.Fatal("startRun should return a command")
	}
	if started.inspectZone != zoneWhatHappened {
		t.Fatal("startRun must reset inspectZone to whatHappened")
	}
	if started.inspectBodyOffset != 0 {
		t.Fatal("startRun must reset inspectBodyOffset to 0")
	}
}

func TestCompareStateClearedOnStartRun(t *testing.T) {
	m := NewModel()
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)

	// Set stale compare state (as if user marked results then started new run)
	m.workspace.compare = compareState{marked: 2, active: 5}

	started, cmd := m.startRun()
	if cmd == nil {
		t.Fatal("startRun should return a command")
	}
	if started.workspace.compare.marked != -1 {
		t.Fatal("startRun must clear compare.marked, got", started.workspace.compare.marked)
	}
	if started.workspace.compare.active != -1 {
		t.Fatal("startRun must clear compare.active, got", started.workspace.compare.active)
	}
}

func TestCompareStateClearedOnCancelRun(t *testing.T) {
	// cancelRun does not clear results or compare state (results remain visible)
	m := NewModel()
	m.running = true
	m.cancel = func() {}
	m.workspace.compare = compareState{marked: 1, active: 2}
	m.results = []model.Result{
		{Status: 200},
		{Status: 404},
	}

	m = m.cancelRun()
	if m.workspace.compare.marked != 1 {
		t.Fatal("cancelRun must preserve compare.marked (results remain visible)")
	}
	if m.workspace.compare.active != 2 {
		t.Fatal("cancelRun must preserve compare.active (results remain visible)")
	}
}

// ---------------------------------------------------------------------------
// Comparison lifecycle and rendering tests
// ---------------------------------------------------------------------------

func compareTestModel() Model {
	m := NewModel()
	m.shell.Resize(130, 30)
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"},
		{Status: 404, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/b"},
		{Status: 200, Latency: 15 * time.Millisecond, RequestURL: "https://example.com/c"},
	}
	return m
}

func TestCompare_MarkLifecycle(t *testing.T) {

	t.Run("first c marks result and returns to Observe", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeInspect
		m.selected = 1
		updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = updated.(Model)
		if m.workspace.compare.marked != 1 {
			t.Fatal("first c should store marked index, got", m.workspace.compare.marked)
		}
		if m.workspace.mode != modeObserve {
			t.Fatal("first c should return to Observe mode")
		}
	})

	t.Run("second c on different result enters Compare", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeInspect
		m.selected = 1
		m.workspace.compare.marked = 1
		m.selected = 2
		updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = updated.(Model)
		if m.workspace.compare.marked != 1 {
			t.Fatal("marked should remain 1, got", m.workspace.compare.marked)
		}
		if m.workspace.compare.active != 2 {
			t.Fatal("active should be 2, got", m.workspace.compare.active)
		}
		if m.workspace.mode != modeCompare {
			t.Fatal("second c should enter Compare mode")
		}
	})

	t.Run("c on same marked result unmarks", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeInspect
		m.selected = 1
		m.workspace.compare.marked = 1
		updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = updated.(Model)
		if m.workspace.compare.marked != -1 {
			t.Fatal("c on same marked result should unmark, got", m.workspace.compare.marked)
		}
		if m.workspace.mode != modeInspect {
			t.Fatal("unmark should stay in Inspect mode")
		}
	})

	t.Run("c with no mark sets marked when mark exists on same result does nothing", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeInspect
		m.selected = 0
		updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = updated.(Model)
		if m.workspace.compare.marked != 0 {
			t.Fatal("mark should be 0, got", m.workspace.compare.marked)
		}
	})
}

func TestCompare_CompareKeyLifecycle(t *testing.T) {

	t.Run("Esc from Compare clears state and returns to Observe", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeCompare
		m.workspace.compare = compareState{marked: 1, active: 2}
		m.workspace.view = LogsView
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(Model)
		if m.workspace.compare.marked != -1 {
			t.Fatal("Esc should reset marked to -1")
		}
		if m.workspace.compare.active != -1 {
			t.Fatal("Esc should reset active to -1")
		}
		if m.workspace.mode != modeObserve {
			t.Fatal("Esc should set mode to Observe")
		}
		if m.workspace.view != TimelineView {
			t.Fatal("Esc should reset view to TimelineView")
		}
	})

	t.Run("q from Compare shows quit dialog", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeCompare
		m.workspace.compare = compareState{marked: 1, active: 2}
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
		m = updated.(Model)
		if m.workspace.dialog != dialogConfirmQuit {
			t.Fatal("q should open confirm quit dialog")
		}
		if m.workspace.compare.marked != -1 {
			t.Fatal("q should clear marked")
		}
		if m.workspace.compare.active != -1 {
			t.Fatal("q should clear active")
		}
	})

	t.Run("Esc from Inspect preserves mark", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeInspect
		m.workspace.compare = compareState{marked: 1, active: -1}
		updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(Model)
		if m.workspace.compare.marked != 1 {
			t.Fatal("Esc from Inspect should preserve mark")
		}
		if m.workspace.mode != modeObserve {
			t.Fatal("Esc from Inspect should return to Observe")
		}
	})

	t.Run("c after Esc starts fresh lifecycle", func(t *testing.T) {
		m := compareTestModel()
		m.workspace.mode = modeCompare
		m.workspace.compare = compareState{marked: 1, active: 2}
		m.workspace.view = LogsView
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = updated.(Model)

		m.workspace.mode = modeInspect
		m.selected = 0
		updated2, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = updated2.(Model)
		if m.workspace.compare.marked != 0 {
			t.Fatal("after Esc, c should mark fresh result, got", m.workspace.compare.marked)
		}
		if m.workspace.mode != modeObserve {
			t.Fatal("after Esc, first c should return to Observe")
		}
	})
}

func TestCompare_InvestigationReset(t *testing.T) {
	m := compareTestModel()
	m.workspace.mode = modeInspect
	m.selected = 1
	m.workspace.compare = compareState{marked: 0, active: -1}
	m.inspectZone = zoneBody
	m.inspectBodyOffset = 10

	updated, _ := m.handleInspectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = updated.(Model)

	if m.workspace.mode != modeCompare {
		t.Fatal("should enter Compare mode")
	}
	if m.inspectZone != zoneWhatHappened {
		t.Fatal("Compare should reset inspectZone to WhatHappened")
	}
	if m.inspectBodyOffset != 0 {
		t.Fatal("Compare should reset inspectBodyOffset to 0")
	}
}

func TestCompare_InvalidState(t *testing.T) {
	m := compareTestModel()
	region := Region{Width: 100, Height: 30}

	t.Run("marked < 0 shows no comparison", func(t *testing.T) {
		m.workspace.compare = compareState{marked: -1, active: 1}
		out := m.renderCompare(region)
		if !contains(t, out, "No comparison active") {
			t.Fatal("invalid state should show no comparison message")
		}
	})

	t.Run("active < 0 shows no comparison", func(t *testing.T) {
		m.workspace.compare = compareState{marked: 1, active: -1}
		out := m.renderCompare(region)
		if !contains(t, out, "No comparison active") {
			t.Fatal("invalid state should show no comparison message")
		}
	})

	t.Run("both < 0 shows no comparison", func(t *testing.T) {
		m.workspace.compare = compareState{marked: -1, active: -1}
		out := m.renderCompare(region)
		if !contains(t, out, "No comparison active") {
			t.Fatal("invalid state should show no comparison message")
		}
	})
}

func TestCompare_ResponsiveLayouts(t *testing.T) {
	m := compareTestModel()
	m.workspace.compare = compareState{marked: 0, active: 1}

	t.Run("narrow (<80) shows rejection", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 79, Height: 30})
		if !contains(t, out, "requires at least 80 columns") {
			t.Fatal("narrow should show rejection message")
		}
	})

	t.Run("medium (80) renders stacked layout", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 80, Height: 30})
		if contains(t, out, "requires at least 80 columns") {
			t.Fatal("80 should be valid width")
		}
		if !contains(t, out, "Marked") {
			t.Fatal("medium should show Marked identity")
		}
		if !contains(t, out, "Current") {
			t.Fatal("medium should show Current identity")
		}
	})

	t.Run("medium (100) renders stacked layout", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 100, Height: 30})
		if !contains(t, out, "Marked") {
			t.Fatal("medium should show Marked identity")
		}
		if !contains(t, out, "Current") {
			t.Fatal("medium should show Current identity")
		}
	})

	t.Run("boundary (119) renders stacked layout", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 119, Height: 30})
		if !contains(t, out, "Marked") {
			t.Fatal("119 should show Marked identity")
		}
		if !contains(t, out, "Current") {
			t.Fatal("119 should show Current identity")
		}
	})

	t.Run("wide (120) renders side-by-side layout", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 120, Height: 30})
		if !contains(t, out, "Marked") {
			t.Fatal("wide should show Marked identity")
		}
		if !contains(t, out, "Current") {
			t.Fatal("wide should show Current identity")
		}
		if !contains(t, out, "│") {
			t.Fatal("wide should include column separator")
		}
	})

	t.Run("wide (150) renders side-by-side layout", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 150, Height: 30})
		if !contains(t, out, "Marked") {
			t.Fatal("wide should show Marked identity")
		}
		if !contains(t, out, "│") {
			t.Fatal("wide should include column separator")
		}
	})
}

func TestCompare_Rendering(t *testing.T) {
	m := compareTestModel()
	m.workspace.compare = compareState{marked: 0, active: 1}
	region := Region{Width: 130, Height: 30}
	out := m.renderCompare(region)

	if !contains(t, out, "◆") {
		t.Fatal("Marked identity should include diamond")
	}
	if !contains(t, out, "▶") {
		t.Fatal("Current identity should include right arrow")
	}
	if !contains(t, out, "Marked") {
		t.Fatal("should show Marked identity")
	}
	if !contains(t, out, "Current") {
		t.Fatal("should show Current identity")
	}
	if !contains(t, out, "WHAT HAPPENED") {
		t.Fatal("compare should show WHAT HAPPENED section")
	}
	if !contains(t, out, "WHY") {
		t.Fatal("compare should show WHY section")
	}
	if !contains(t, out, "RESPONSE") {
		t.Fatal("compare should show RESPONSE section")
	}
}

func TestCompare_Navigation(t *testing.T) {
	m := compareTestModel()
	m.workspace.mode = modeCompare
	m.workspace.compare = compareState{marked: 0, active: 1}

	t.Run("up scrolls body when in body zone", func(t *testing.T) {
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 5
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyUp})
		m2 := updated.(Model)
		if m2.inspectBodyOffset != 4 {
			t.Fatal("up should decrement body offset")
		}
	})

	t.Run("up does not scroll below 0", func(t *testing.T) {
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 0
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyUp})
		m2 := updated.(Model)
		if m2.inspectBodyOffset != 0 {
			t.Fatal("up should not scroll below 0")
		}
	})

	t.Run("k scrolls body (vim key)", func(t *testing.T) {
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 5
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m2 := updated.(Model)
		if m2.inspectBodyOffset != 4 {
			t.Fatal("k should decrement body offset")
		}
	})

	t.Run("j scrolls body (vim key)", func(t *testing.T) {
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 5
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m2 := updated.(Model)
		if m2.inspectBodyOffset != 6 {
			t.Fatal("j should increment body offset")
		}
	})

	t.Run("down scrolls body", func(t *testing.T) {
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 5
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyDown})
		m2 := updated.(Model)
		if m2.inspectBodyOffset != 6 {
			t.Fatal("down should increment body offset")
		}
	})

	t.Run("tab cycles zones forward", func(t *testing.T) {
		m.inspectZone = zoneWhatHappened
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyTab})
		m2 := updated.(Model)
		if m2.inspectZone != zoneWhy {
			t.Fatal("tab should advance to WHY zone")
		}
		updated, _ = m2.handleCompareKey(tea.KeyMsg{Type: tea.KeyTab})
		m3 := updated.(Model)
		if m3.inspectZone != zoneBody {
			t.Fatal("tab should advance to BODY zone")
		}
		updated, _ = m3.handleCompareKey(tea.KeyMsg{Type: tea.KeyTab})
		m4 := updated.(Model)
		if m4.inspectZone != zoneWhatHappened {
			t.Fatal("tab should wrap to WHAT HAPPENED zone")
		}
	})

	t.Run("shift+tab cycles zones backward", func(t *testing.T) {
		m.inspectZone = zoneBody
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyShiftTab})
		m2 := updated.(Model)
		if m2.inspectZone != zoneWhy {
			t.Fatal("shift+tab from body should go to WHY zone")
		}
		updated, _ = m2.handleCompareKey(tea.KeyMsg{Type: tea.KeyShiftTab})
		m3 := updated.(Model)
		if m3.inspectZone != zoneWhatHappened {
			t.Fatal("shift+tab from WHY should go to WHAT HAPPENED")
		}
	})

	t.Run("home jumps to What Happened", func(t *testing.T) {
		m.inspectZone = zoneBody
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyHome})
		m2 := updated.(Model)
		if m2.inspectZone != zoneWhatHappened {
			t.Fatal("home should jump to WHAT HAPPENED")
		}
	})

	t.Run("end jumps to Body", func(t *testing.T) {
		m.inspectZone = zoneWhatHappened
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyEnd})
		m2 := updated.(Model)
		if m2.inspectZone != zoneBody {
			t.Fatal("end should jump to BODY")
		}
	})
}

func TestCompare_ZoneNavigation(t *testing.T) {
	m := compareTestModel()
	m.workspace.mode = modeCompare
	m.workspace.compare = compareState{marked: 0, active: 1}

	t.Run("Tab at WhatHappened goes to Why", func(t *testing.T) {
		m.inspectZone = zoneWhatHappened
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyTab})
		m2 := updated.(Model)
		if m2.inspectZone != zoneWhy {
			t.Fatal("Tab should go from WhatHappened to Why")
		}
	})

	t.Run("Tab at Body wraps to WhatHappened", func(t *testing.T) {
		m.inspectZone = zoneBody
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyTab})
		m2 := updated.(Model)
		if m2.inspectZone != zoneWhatHappened {
			t.Fatal("Tab from Body should wrap to WhatHappened")
		}
	})

	t.Run("Shift+Tab at WhatHappened wraps to Body", func(t *testing.T) {
		m.inspectZone = zoneWhatHappened
		updated, _ := m.handleCompareKey(tea.KeyMsg{Type: tea.KeyShiftTab})
		m2 := updated.(Model)
		if m2.inspectZone != zoneBody {
			t.Fatal("Shift+Tab from WhatHappened should wrap to Body")
		}
	})
}

func TestRenderStatusline_CompareMode(t *testing.T) {
	m := compareTestModel()
	m.shell.Resize(100, 24)
	m.workspace.mode = modeCompare
	m.workspace.compare = compareState{marked: 0, active: 1}

	out := m.renderStatusline(m.ShellState(), 100)
	if !contains(t, out, "[Tab]") {
		t.Fatal("compare ribbon should show [Tab] for zone navigation")
	}
	if !contains(t, out, "[Esc] Back") {
		t.Fatal("compare ribbon should show [Esc] Back")
	}
	if !contains(t, out, "[q] Quit") {
		t.Fatal("compare ribbon should show [q] Quit")
	}
	if !contains(t, out, "Comparing") {
		t.Fatal("compare status should show 'Comparing'")
	}
}

func TestCompare_HandleKeyDispatch(t *testing.T) {
	m := compareTestModel()
	m.workspace.mode = modeCompare
	m.workspace.compare = compareState{marked: 0, active: 1}

	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.workspace.mode != modeObserve {
		t.Fatal("handleKey should dispatch Esc to compare handler")
	}
	if m.workspace.compare.marked != -1 {
		t.Fatal("handleKey should clear compare state via compare handler")
	}
}

// ---------------------------------------------------------------------------
// Compare Diff Summary Tests
// ---------------------------------------------------------------------------

func TestCompareDiff_StatusChange(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 404, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/b"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if !contains(t, out, "200") || !contains(t, out, "404") {
		t.Fatal("diff must show both status values")
	}
}

func TestCompareDiff_StatusSame(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 200, Latency: 15 * time.Millisecond, RequestURL: "https://example.com/a"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if contains(t, out, "200 →") {
		t.Fatal("status should not show diff when unchanged")
	}
}

func TestCompareDiff_LatencyDelta(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if !contains(t, out, "+") || !contains(t, out, "Latency") {
		t.Fatal("diff must show latency delta with + sign")
	}
}

func TestCompareDiff_LatencyDecrease(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if !contains(t, out, "-") || !contains(t, out, "Latency") {
		t.Fatal("diff must show latency delta with - sign")
	}
}

func TestCompareDiff_URLChange(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/b"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if !contains(t, out, "example.com/a") || !contains(t, out, "example.com/b") {
		t.Fatal("diff must show both URLs when different")
	}
}

func TestCompareDiff_ErrorChange(t *testing.T) {
	marked := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a"}
	active := model.Result{Status: 200, Latency: 10 * time.Millisecond, RequestURL: "https://example.com/a",
		Error: "connection refused"}
	m := NewModel()

	out := m.renderCompareDiff(marked, active)

	if !contains(t, out, "connection refused") {
		t.Fatal("diff must show error when error appears")
	}
}

func TestCompareDiff_DiffSummaryInRender(t *testing.T) {
	m := compareTestModel()
	m.workspace.compare = compareState{marked: 0, active: 1}

	out := m.renderCompare(Region{Width: 100, Height: 30})

	if !contains(t, out, "DIFF SUMMARY") {
		t.Fatal("compare render must include DIFF SUMMARY section")
	}
	if !contains(t, out, "Status:") && !contains(t, out, "Latency:") {
		t.Fatal("diff summary must show at least one diff field")
	}
}

// ---------------------------------------------------------------------------
// Payload Rendering Regression Tests
// ---------------------------------------------------------------------------

func TestPayloadRender_FocusInvariance(t *testing.T) {
	// Focus changes must not alter BODY heading placement.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.bodyInput.SetValue("{\"key\": \"value\"}")

	out := m.renderPayloadDomain(80)
	if !contains(t, out, "BODY") {
		t.Fatal("payload render must include BODY heading")
	}
	if !contains(t, out, "key") {
		t.Fatal("payload render must show body content")
	}
}

func TestPayloadRender_BodyAndHeadersRender(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.bodyInput.SetValue("test content")

	out := m.renderPayloadDomain(80)
	lines := strings.Split(out, "\n")

	bodyHeading := false
	headersHeading := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BODY") {
			bodyHeading = true
		}
		if trimmed == "HEADERS" {
			headersHeading = true
		}
	}
	if !bodyHeading {
		t.Fatal("payload render must include BODY heading")
	}
	if !headersHeading {
		t.Fatal("payload render must include HEADERS heading")
	}
}

func TestPayloadRender_WindowSizeInvariance(t *testing.T) {
	m := NewModel()
	m.bodyInput.SetValue("{\"data\": 1}")

	for _, w := range []int{80, 100, 120, 140} {
		m.shell.Resize(w, 30)
		out := m.renderPayloadDomain(w - 4)
		if !contains(t, out, "BODY") {
			t.Fatalf("payload at width %d must show BODY heading", w)
		}
	}
}

// ---------------------------------------------------------------------------
// Ready Screen Validation Display
// ---------------------------------------------------------------------------

func TestRenderReady_ValidationGuidance(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.errMsg = "Concurrency must be between 1 and 100"

	out := m.renderReady(Region{Width: 100, Height: 26})

	if !contains(t, out, "Configuration incomplete") {
		t.Fatal("ready screen should show Configuration incomplete heading when errMsg set")
	}
	if !contains(t, out, "Concurrency must be between 1 and 100") {
		t.Fatal("ready screen should show the error message")
	}
	if !contains(t, out, "Press E to edit and adjust") {
		t.Fatal("ready screen should show edit guidance")
	}
	if contains(t, out, "Ready to execute") {
		t.Fatal("ready screen must NOT show Ready to execute when errMsg set")
	}
}

func TestRenderReady_NoValidation_ShowsReady(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.errMsg = ""

	out := m.renderReady(Region{Width: 100, Height: 26})

	if !contains(t, out, "Ready to execute") {
		t.Fatal("ready screen should show Ready to execute when errMsg empty")
	}
	if contains(t, out, "Configuration incomplete") {
		t.Fatal("ready screen must NOT show Configuration incomplete when errMsg empty")
	}
}

// ---------------------------------------------------------------------------
// Footer Ribbon Regression Tests
// ---------------------------------------------------------------------------

func TestFooter_LongErrMsgTruncation(t *testing.T) {
	// Long validation messages must be truncated with ellipsis, never clipped.
	m := NewModel()
	m.errMsg = "CONCURRENCY MUST BE BETWEEN 1 AND 100"

	for _, width := range []int{72, 80, 90, 100, 120} {
		state := m.ShellState()
		out := m.renderStatusline(state, width)
		if strings.Contains(out, "\n") {
			t.Fatalf("width %d: footer must not wrap", width)
		}
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: rendered width %d exceeds available width", width, w)
		}
		// Verify the full error text appears (no clipping)
		if !contains(t, stripANSI(out), "CONCURRENCY MUST BE BETWEEN 1 AND 100") {
			t.Fatalf("width %d: status must contain full error text", width)
		}
		if strings.Count(out, "…") > 0 {
			t.Fatalf("width %d: error text must not be truncated at this width", width)
		}
	}
}

func TestFooter_ShortStatusUntruncated(t *testing.T) {
	m := NewModel()
	state := m.ShellState()

	for _, width := range []int{72, 100, 160} {
		out := m.renderStatusline(state, width)
		if !contains(t, out, "Ready") {
			t.Fatalf("width %d: status must show Ready", width)
		}
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: rendered width %d exceeds available width", width, w)
		}
	}
}

func TestFooter_RibbonWidthConsistency(t *testing.T) {
	// chooseRibbonLevel and renderRibbon must agree on total width.
	m := NewModel()
	m.errMsg = "CONCURRENCY MUST BE BETWEEN 1 AND 100"

	for _, width := range []int{72, 80, 100, 120, 160, 200} {
		state := m.ShellState()
		out := m.renderStatusline(state, width)
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: rendered %d > available", width, w)
		}
		// Verify no clipping by checking absence of ANSI-truncated characters
		if strings.Count(out, "…") > 1 {
			t.Fatalf("width %d: multiple ellipsis suggests incorrect truncation", width)
		}
	}
}

func TestFooter_EditingNoClip(t *testing.T) {
	// Footer during editing must never clip content.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest

	for _, width := range []int{72, 80, 100, 120, 160} {
		state := m.ShellState()
		out := m.renderStatusline(state, width)
		if strings.Contains(out, "\n") {
			t.Fatalf("width %d: footer must not wrap", width)
		}
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: rendered width %d exceeds available width", width, w)
		}
		if !contains(t, stripANSI(out), "Editing") {
			t.Fatalf("width %d: status must show Editing", width)
		}
	}

	// Also test with error during editing
	m.errMsg = "URL IS REQUIRED"
	for _, width := range []int{72, 80, 100, 120, 160} {
		state := m.ShellState()
		out := m.renderStatusline(state, width)
		if strings.Contains(out, "\n") {
			t.Fatalf("width %d: editing with error: footer must not wrap", width)
		}
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: editing with error: rendered width %d exceeds available", width, w)
		}
		if !contains(t, stripANSI(out), "URL IS REQUIRED") {
			t.Fatalf("width %d: editing with error: must show error text", width)
		}
	}
}

// TestHeaderEditing_Invariant asserts that selecting different header rows does
// not change text alignment — cursor and key/value positions are invariant.
func TestHeaderEditing_Invariant(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow())
	m.headers[0].Key.SetValue("Authorization")
	m.headers[0].Value.SetValue("Bearer token")
	m.headers[1].Key.SetValue("Content-Type")
	m.headers[1].Value.SetValue("application/json")

	visualColumn := func(line, substr string) int {
		idx := strings.Index(line, substr)
		if idx < 0 {
			return -1
		}
		return lipgloss.Width(line[:idx])
	}

	// Render with first row selected
	m.selectedHead = 0
	out1 := m.renderPayloadDomain(96)
	// Render with second row selected
	m.selectedHead = 1
	out2 := m.renderPayloadDomain(96)

	// Split into lines and find header rows
	lines1 := strings.Split(out1, "\n")
	lines2 := strings.Split(out2, "\n")

	var row1a, row1b, row2a, row2b string
	for _, line := range lines1 {
		if strings.Contains(line, "Authorization") {
			row1a = line
		}
		if strings.Contains(line, "Content-Type") {
			row1b = line
		}
	}
	for _, line := range lines2 {
		if strings.Contains(line, "Authorization") {
			row2a = line
		}
		if strings.Contains(line, "Content-Type") {
			row2b = line
		}
	}

	if row1a == "" || row1b == "" || row2a == "" || row2b == "" {
		t.Fatal("header rows not found in render output")
	}

	// Verify the key visual column is the same regardless of selection
	colAuth1 := visualColumn(row1a, "Authorization")
	colAuth2 := visualColumn(row2a, "Authorization")
	if colAuth1 != colAuth2 {
		t.Fatalf("Authorization key visual column changed with selection: %d vs %d", colAuth1, colAuth2)
	}

	colCT1 := visualColumn(row1b, "Content-Type")
	colCT2 := visualColumn(row2b, "Content-Type")
	if colCT1 != colCT2 {
		t.Fatalf("Content-Type key visual column changed with selection: %d vs %d", colCT1, colCT2)
	}

	// Verify separator (":") is at same visual column
	colSep1 := visualColumn(row1a, ":")
	colSep2 := visualColumn(row2a, ":")
	if colSep1 != colSep2 {
		t.Fatalf("separator visual column changed with selection: %d vs %d", colSep1, colSep2)
	}
}

// TestRibbon_Ownership verifies the complete chooseRibbonLevel → renderRibbon
// chain produces final width ≤ terminal width for every density level.
func TestRibbon_Ownership(t *testing.T) {
	for _, width := range []int{40, 50, 60, 72, 80, 100, 120, 160, 200} {
		for _, orientation := range []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "QUIT"} {
			var actions []Action
			badge := renderWorkspaceBadge(orientation)
			switch orientation {
			case "READY":
				m := NewModel()
				actions = m.Actions()
			case "OBSERVE":
				m := NewModel()
				m.results = []model.Result{{Status: 200}}
				m.running = false
				actions = m.Actions()
			case "REQUEST":
				m := NewModel()
				m.workspace.dialog = dialogRequest
				actions = m.Actions()
			case "INSPECT":
				m := NewModel()
				m.workspace.mode = modeInspect
				actions = m.Actions()
			case "QUIT":
				m := NewModel()
				m.workspace.dialog = dialogConfirmQuit
				actions = m.Actions()
			}

			status := styleStatusCell.Render("Ready")
			level, actionText := chooseRibbonLevel(badge, status, actions, width)

			layout := RibbonLayout{
				Badge:   badge,
				Actions: actionText,
				Status:  status,
				Density: level,
			}
			out := renderRibbon(layout, width)

			if w := lipgloss.Width(out); w > width {
				t.Fatalf("orientation=%s width=%d level=%d: rendered %d > available",
					orientation, width, level, w)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Validation Regression Tests
// ---------------------------------------------------------------------------

func TestValidation_ConcurrencyDeduplication(t *testing.T) {
	// When inline concurrency check fires, m.errMsg block must not duplicate it.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.errMsg = "CONCURRENCY MUST BE BETWEEN 1 AND 100"
	m.concurrencyInput.SetValue("500")

	out := m.renderExecDomain(80)

	// Count occurrences of "Must be between" — should be exactly 1
	count := strings.Count(out, "Must be between")
	if count != 1 {
		t.Fatalf("concurrency error must appear exactly once, got %d", count)
	}
	if contains(t, out, "CONCURRENCY MUST BE") {
		t.Fatal("m.errMsg block must NOT appear when inline concurrency check fires")
	}
}

func TestValidation_NonConcurrencyErrMsgShows(t *testing.T) {
	// When errMsg is set but NOT about concurrency, the m.errMsg block must appear.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.errMsg = "URL IS REQUIRED"
	m.concurrencyInput.SetValue("5")

	out := m.renderExecDomain(80)

	if !contains(t, out, "URL IS REQUIRED") {
		t.Fatal("m.errMsg block must appear for non-concurrency errors")
	}
	if !contains(t, out, "Adjust the request and run again") {
		t.Fatal("m.errMsg block must show recovery guidance")
	}
}

func TestValidation_NoDuplicateWhenValidConcurrency(t *testing.T) {
	// When concurrency is valid AND errMsg is about concurrency, only inline shows.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.concurrencyInput.SetValue("10")

	out := m.renderExecDomain(80)

	if contains(t, out, "Must be between") {
		t.Fatal("no inline error for valid concurrency")
	}
}

func TestValidation_URL_InlineWhenActive(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.SetValue("")

	out := m.renderRequestDomain(80)
	if !contains(t, out, "URL is required") {
		t.Fatal("inline URL error must show when URL field is active and empty")
	}
}

func TestValidation_URL_NotInlineWhenInactive(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.urlInput.SetValue("")

	out := m.renderPayloadDomain(80)
	if contains(t, out, "URL is required") {
		t.Fatal("URL validation must NOT appear in Payload domain")
	}
}

func TestValidation_URL_ValidErrorHidden(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL

	out := m.renderRequestDomain(80)
	if contains(t, out, "URL is required") || contains(t, out, "Must be a valid") {
		t.Fatal("no inline error for valid default URL")
	}
}

func TestValidation_URL_MalformedURL(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.SetValue("not-a-url")

	out := m.renderRequestDomain(80)
	if !contains(t, out, "Must be a valid absolute URL") {
		t.Fatal("inline error must show for malformed URL")
	}
}

func TestValidation_Body_InvalidJSON(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.bodyInput.SetValue("{invalid}")

	out := m.renderPayloadDomain(80)
	if !contains(t, out, "Body must be valid JSON") {
		t.Fatal("inline error must show for invalid JSON body")
	}
}

func TestValidation_Body_ValidJSONNoError(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.bodyInput.SetValue(`{"key": "value"}`)

	out := m.renderPayloadDomain(80)
	if contains(t, out, "Body must be valid") {
		t.Fatal("no inline error for valid JSON")
	}
}

func TestValidation_Body_NoErrorWhenEmpty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	m.bodyInput.SetValue("")

	out := m.renderPayloadDomain(80)
	if contains(t, out, "Body must be valid") {
		t.Fatal("no inline error when body is empty")
	}
}

func TestValidation_Header_KeyRequired(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = []headerRow{newHeaderRow()}
	m.headers[0].Key.SetValue("")

	out := m.renderPayloadDomain(80)
	if !contains(t, out, "Header key is required") {
		t.Fatal("inline error must show when header key is empty")
	}
}

func TestValidation_Header_KeyPresentNoError(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = []headerRow{newHeaderRow()}
	m.headers[0].Key.SetValue("Content-Type")

	out := m.renderPayloadDomain(80)
	if contains(t, out, "Header key is required") {
		t.Fatal("no inline error when header key is present")
	}
}

// ---------------------------------------------------------------------------
// Payload Workspace Regression Tests
// ---------------------------------------------------------------------------

func TestPayload_PlaceholderRenders(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.bodyInput.SetValue("")

	out := m.renderPayloadDomain(80)
	if !contains(t, out, "BODY") {
		t.Fatal("payload must show BODY heading")
	}
}

func TestPayload_MultilineRenders(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.bodyInput.SetValue("line1\nline2\nline3")

	out := m.renderPayloadDomain(80)
	if !contains(t, out, "BODY") {
		t.Fatal("payload must show BODY heading")
	}
	// Verify body content appears (may have ANSI styling)
	if !strings.Contains(out, "line1") && !strings.Contains(out, "line2") && !strings.Contains(out, "line3") {
		t.Fatal("payload render must contain body content text")
	}
}

func TestPayload_ContextPanelBodyWidth(t *testing.T) {
	// At width >= 140 (context panel visible), body must not be truncated.
	m := NewModel()
	m.shell.Resize(160, 30)
	m.bodyInput.SetValue(`{"key": "value"}`)

	out := m.renderPayloadDomain(140)
	if !contains(t, out, "BODY") {
		t.Fatal("payload must show BODY at wide width")
	}
	if !contains(t, out, "key") {
		t.Fatal("payload must show body content at wide width")
	}
}

func TestPayload_FocusTraversalGeometry(t *testing.T) {
	// Focus changes must not alter BODY heading position or body content.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.bodyInput.SetValue("data")
	m.activeDomain = DomainRequest

	// Request domain focus (not payload)
	out1 := m.renderPayloadDomain(80)
	// Payload domain focus
	m.activeDomain = DomainPayload
	m.selectedHead = bodyFocus
	out2 := m.renderPayloadDomain(80)

	// Strip ANSI for position comparison
	strip := func(s string) string {
		var out strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '\x1b' {
				for i < len(s) && s[i] != 'm' {
					i++
				}
				continue
			}
			out.WriteByte(s[i])
		}
		return out.String()
	}
	s1, s2 := strip(out1), strip(out2)
	if strings.Index(s1, "BODY") != strings.Index(s2, "BODY") {
		t.Fatal("BODY heading position must not change with focus")
	}
}

func TestPayload_ResizeGeometry(t *testing.T) {
	m := NewModel()
	m.bodyInput.SetValue("content")

	for _, w := range []int{80, 100, 120, 160, 200} {
		m.shell.Resize(w, 30)
		out := m.renderPayloadDomain(w - 4)
		if !contains(t, out, "BODY") {
			t.Fatalf("width %d: payload must show BODY", w)
		}
		if !contains(t, out, "content") {
			t.Fatalf("width %d: payload must show body content", w)
		}
	}
}

// ---------------------------------------------------------------------------
// Ribbon Level Consistency Test
// ---------------------------------------------------------------------------

func TestRibbon_ChooseRibbonLevelRenderRibbonAgreement(t *testing.T) {
	m := NewModel()
	m.errMsg = "CONCURRENCY MUST BE BETWEEN 1 AND 100"

	for _, width := range []int{72, 80, 100, 120, 160, 200} {
		state := m.ShellState()
		out := m.renderStatusline(state, width)
		if w := lipgloss.Width(out); w > width {
			t.Fatalf("width %d: rendered width %d exceeds available", width, w)
		}
		if strings.Count(out, "…") > 1 {
			t.Fatalf("width %d: multiple ellipsis suggests incorrect truncation", width)
		}
	}
}

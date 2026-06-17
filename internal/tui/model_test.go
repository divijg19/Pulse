package tui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/divijg19/Pulse/internal/model"
)

func TestFocusMovement(t *testing.T) {
	m := NewModel()
	if m.focus != focusURL {
		t.Fatalf("initial focus = %v", m.focus)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focus != focusConcurrency {
		t.Fatalf("focus after tab = %v", m.focus)
	}
}

func TestMethodSelection(t *testing.T) {
	m := NewModel()
	m.focus = focusMethod

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if got := m.methodIndex; got != 1 {
		t.Fatalf("method index = %d", got)
	}
}

func TestConcurrencyClamping(t *testing.T) {
	m := NewModel()
	m.setConcurrency(500)
	if got := m.concurrency(); got != 100 {
		t.Fatalf("concurrency high clamp = %d", got)
	}
	m.setConcurrency(-10)
	if got := m.concurrency(); got != 1 {
		t.Fatalf("concurrency low clamp = %d", got)
	}
}

func TestPayloadEditorState(t *testing.T) {
	m := NewModel()
	m.focus = focusPayload

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.showPayload {
		t.Fatal("payload editor should be visible")
	}
	if len(m.headers) != 1 {
		t.Fatalf("headers len = %d", len(m.headers))
	}
}

func TestRunStartAndCancelTransitions(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	m := NewModel()
	m.urlInput.SetValue(upstream.URL)
	m.setConcurrency(1)

	started, cmd := m.startRun()
	if cmd == nil {
		t.Fatal("startRun should return a command")
	}
	if !started.running {
		t.Fatal("model should be running")
	}

	cancelled := started.cancelRun()
	if cancelled.status != "CANCELLED" {
		t.Fatalf("status = %q", cancelled.status)
	}
}

func TestTabSwitching(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(Model)
	if m.activeTab != tabLogs {
		t.Fatalf("active tab = %v", m.activeTab)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(Model)
	if m.activeTab != tabTimeline {
		t.Fatalf("active tab = %v", m.activeTab)
	}
}

func TestResultSelectionAndInspector(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.results = []model.Result{
		{Status: http.StatusOK},
		{Status: http.StatusInternalServerError},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 1 {
		t.Fatalf("selected = %d", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.inspector {
		t.Fatal("inspector should be open")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.inspector {
		t.Fatal("inspector should close on esc")
	}
}

func TestHeaderAddRemove(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders

	// handleHeaderKey auto-adds one header when the slice is empty,
	// so after the first ctrl+n there will be 2.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = updated.(Model)
	if len(m.headers) != 2 {
		t.Fatalf("after ctrl+n headers = %d (expected 2)", len(m.headers))
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)
	if len(m.headers) != 1 {
		t.Fatalf("after ctrl+d headers = %d (expected 1)", len(m.headers))
	}
}

func TestRejectsOversizedBody(t *testing.T) {
	m := NewModel()
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)
	m.bodyInput.CharLimit = maxTUIBodyBytes + 100
	m.bodyInput.SetValue(strings.Repeat("x", maxTUIBodyBytes+1))

	started, cmd := m.startRun()
	if cmd != nil {
		t.Fatal("startRun should return nil cmd when body too large")
	}
	if started.running {
		t.Fatal("model should not be running with oversized body")
	}
	if started.status != "BODY TOO LARGE (MAX 1MB)" {
		t.Fatalf("status = %q", started.status)
	}
}

func TestAutoScrollToggle(t *testing.T) {
	m := NewModel()
	if !m.autoScroll {
		t.Fatal("autoScroll should default to true")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m = updated.(Model)
	if m.autoScroll {
		t.Fatal("autoScroll should toggle to false")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m = updated.(Model)
	if !m.autoScroll {
		t.Fatal("autoScroll should toggle back to true")
	}

	m.autoScroll = false
	m.urlInput.SetValue("https://example.com/api")
	m, _ = m.startRun()
	if !m.autoScroll {
		t.Fatal("autoScroll should reset to true on startRun")
	}
}

func TestAutoScrollSelectionAdvances(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.autoScroll = true
	m.startRun()

	msg := resultMsg{Result: model.Result{Status: 200, Latency: 10 * time.Millisecond}}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.selected != 0 {
		t.Fatalf("after first result, selected = %d (expected 0)", m.selected)
	}

	msg2 := resultMsg{Result: model.Result{Status: 200, Latency: 20 * time.Millisecond}}
	updated2, _ := m.Update(msg2)
	m = updated2.(Model)

	if m.selected != 1 {
		t.Fatalf("after second result with auto-scroll, selected = %d (expected 1)", m.selected)
	}
}

func TestAutoScrollDoesNotAdvanceWhenScrolledAway(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.autoScroll = true
	m.results = []model.Result{
		{Status: 200},
		{Status: 200},
		{Status: 200},
	}
	m.selected = 0

	msg := resultMsg{Result: model.Result{Status: 200, Latency: 5 * time.Millisecond}}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.selected != 0 {
		t.Fatalf("selected should stay at 0 when scrolled away, got %d", m.selected)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("width/height = %d/%d", m.width, m.height)
	}
}

func TestTickMsg_Idle(t *testing.T) {
	m := NewModel()
	m.running = false
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(Model)
	if m.elapsed != 0 {
		t.Fatal("elapsed should not update when not running")
	}
}

func TestResultMsg_AutoScrollOffKeepsSelection(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.autoScroll = false
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
		{Status: 200, Latency: 10 * time.Millisecond},
	}
	m.selected = 0

	msg := resultMsg{Result: model.Result{Status: 200, Latency: 5 * time.Millisecond}}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.selected != 0 {
		t.Fatalf("selected should stay at 0 with auto-scroll off, got %d", m.selected)
	}
}

func TestResultMsg_CapEnforced(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.autoScroll = false

	capResults := 10000
	for i := 0; i < capResults+5; i++ {
		msg := resultMsg{Result: model.Result{Status: 200, Latency: 1 * time.Millisecond}}
		updated, _ := m.Update(msg)
		m = updated.(Model)
	}

	if len(m.results) != capResults {
		t.Fatalf("results capped at %d, got %d", capResults, len(m.results))
	}
}

func TestRunFinishedMsg(t *testing.T) {
	m := NewModel()
	m.running = true
	m.startedAt = time.Now().Add(-2 * time.Second)
	m.status = "RUNNING"

	updated, _ := m.Update(runFinishedMsg{})
	m = updated.(Model)

	if m.running {
		t.Fatal("running should be false after runFinishedMsg")
	}
	if m.status != "COMPLETE" {
		t.Fatalf("status = %q (expected COMPLETE)", m.status)
	}
}

func TestCtrlC_WhileRunning(t *testing.T) {
	m := NewModel()
	m.running = true
	m.cancel = func() {}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("ctrl+c should return a quit command")
	}
}

func TestCtrlC_Idle(t *testing.T) {
	m := NewModel()
	m.running = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("ctrl+c while idle should return a quit command")
	}
}

func TestCtrlR_WhileRunning(t *testing.T) {
	m := NewModel()
	m.running = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = updated.(Model)

	if cmd != nil {
		t.Fatal("ctrl+r while running should be a no-op (nil cmd)")
	}
}

func TestCtrlX_Cancel(t *testing.T) {
	m := NewModel()
	m.running = true
	m.cancel = func() {}
	m.status = "RUNNING"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m = updated.(Model)

	if m.status != "CANCELLED" {
		t.Fatalf("status = %q (expected CANCELLED)", m.status)
	}
}

func TestEsc_FromBodyFocus(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusBody
	m.headers = append(m.headers, newHeaderRow())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.focus != focusHeaders {
		t.Fatalf("focus = %v (expected focusHeaders)", m.focus)
	}
}

func TestEsc_FromInspector(t *testing.T) {
	m := NewModel()
	m.inspector = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.inspector {
		t.Fatal("esc should close the inspector")
	}
}

func TestInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return a non-nil command")
	}
}

func TestConcurrency_ParseInvalid(t *testing.T) {
	m := NewModel()
	m.ccInput.SetValue("abc")
	if got := m.concurrency(); got != 10 {
		t.Fatalf("invalid concurrency should fallback to default, got %d", got)
	}

	m.ccInput.SetValue("")
	if got := m.concurrency(); got != 10 {
		t.Fatalf("empty concurrency should fallback to default, got %d", got)
	}

	m.ccInput.SetValue("7")
	if got := m.concurrency(); got != 7 {
		t.Fatalf("valid concurrency should be 7, got %d", got)
	}
}

func TestFormatDuration(t *testing.T) {
	tt := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0.00s"},
		{-1 * time.Nanosecond, "0.00s"},
		{500 * time.Millisecond, "0.50s"},
		{1 * time.Second, "1.00s"},
		{90 * time.Second, "1m 30s"},
		{120 * time.Second, "2m 0s"},
		{150 * time.Second, "2m 30s"},
	}

	for _, tc := range tt {
		got := formatDuration(tc.duration)
		if got != tc.expected {
			t.Errorf("formatDuration(%v) = %q (expected %q)", tc.duration, got, tc.expected)
		}
	}
}

func TestHeaderMap(t *testing.T) {
	m := NewModel()
	m.headers = append(m.headers, headerRow{})
	m.headers[0].Key.SetValue("Content-Type")
	m.headers[0].Value.SetValue("application/json")
	m.headers = append(m.headers, headerRow{})
	m.headers[1].Key.SetValue("")
	m.headers[1].Value.SetValue("should-skip")

	hm := m.headerMap()
	if len(hm) != 1 {
		t.Fatalf("headerMap should have 1 entry (empty key skipped), got %d", len(hm))
	}
	if hm["Content-Type"] != "application/json" {
		t.Fatalf("Content-Type = %q", hm["Content-Type"])
	}
}

func TestClamp(t *testing.T) {
	if got := clamp(5, 0, 10); got != 5 {
		t.Fatalf("clamp(5, 0, 10) = %d", got)
	}
	if got := clamp(-5, 0, 10); got != 0 {
		t.Fatalf("clamp(-5, 0, 10) = %d", got)
	}
	if got := clamp(15, 0, 10); got != 10 {
		t.Fatalf("clamp(15, 0, 10) = %d", got)
	}
}

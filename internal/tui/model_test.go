package tui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
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

func TestResultSelection_PageUpDown(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.height = 30
	for i := 0; i < 50; i++ {
		m.results = append(m.results, model.Result{Status: 200})
	}
	m.selected = 25

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(Model)
	expected := 25 + max(5, 30/2)
	if m.selected != expected {
		t.Fatalf("pgdown: selected = %d, want %d", m.selected, expected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updated.(Model)
	expected = 25
	if m.selected != expected {
		t.Fatalf("pgup: selected = %d, want %d", m.selected, expected)
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
	initialWidth := m.bodyInput.Width()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("width/height = %d/%d", m.width, m.height)
	}
	if got := m.bodyInput.Width(); got <= initialWidth {
		t.Fatalf("bodyInput.Width() = %d after resize to 120, expected > initial %d", got, initialWidth)
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

	// First ctrl+c shows confirmation, does not quit
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)

	if cmd != nil {
		t.Fatal("first ctrl+c while running should return nil cmd (confirmation)")
	}
	if !m.confirmQuit {
		t.Fatal("confirmQuit should be set after first ctrl+c")
	}

	// Second ctrl+c confirms quit
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("second ctrl+c should return a quit command")
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
	if m.running {
		t.Fatal("running should be false after cancelRun")
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

func TestStartupMsg_SetsDefaults(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(startupMsg{})
	m = updated.(Model)
	if m.width != 80 {
		t.Fatalf("width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Fatalf("height = %d, want 24", m.height)
	}
}

func TestStartupMsg_AfterWindowSize(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)
	updated, _ = m.Update(startupMsg{})
	m = updated.(Model)
	if m.width != 120 {
		t.Fatal("startupMsg should not override existing width")
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

func TestMoveFocus_NoPayload(t *testing.T) {
	m := NewModel()
	m.showPayload = false
	m.focus = focusMethod

	expected := []focusTarget{focusMethod, focusURL, focusConcurrency, focusPayload, focusResults}
	for i, want := range expected {
		if m.focus != want {
			t.Fatalf("step %d: expected focus %v, got %v", i, want, m.focus)
		}
		if want == focusResults {
			break
		}
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(Model)
	}
}

func TestMoveFocus_ShiftTabWrap(t *testing.T) {
	m := NewModel()
	m.focus = focusMethod

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.focus != focusResults {
		t.Fatalf("shift+tab from method should wrap to results, got %v", m.focus)
	}
}

func TestMoveFocus_WithPayload(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.headers = append(m.headers, newHeaderRow())
	m.focus = focusMethod

	expected := []focusTarget{
		focusMethod, focusURL, focusConcurrency, focusPayload,
		focusHeaders, focusBody, focusResults,
	}

	for i, want := range expected {
		if m.focus != want {
			t.Fatalf("step %d: expected focus %v, got %v", i, want, m.focus)
		}
		if want == focusResults {
			break
		}
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(Model)
	}
}

func TestUpdateFocusedInput_NonInputState(t *testing.T) {
	m := NewModel()

	m.focus = focusMethod
	updated, cmd := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m2 := updated.(Model)
	if cmd != nil {
		t.Fatal("updateFocusedInput should return nil cmd for focusMethod")
	}
	if m2.focus != focusMethod {
		t.Fatal("focus should remain unchanged")
	}

	m.focus = focusPayload
	_, cmd2 := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd2 != nil {
		t.Fatal("updateFocusedInput should return nil cmd for focusPayload")
	}

	m.focus = focusResults
	_, cmd3 := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd3 != nil {
		t.Fatal("updateFocusedInput should return nil cmd for focusResults")
	}
}

func TestFocusMethod_WrapAtBoundaries(t *testing.T) {
	m := NewModel()
	m.focus = focusMethod
	methods := runconfig.AllowedMethods()
	last := len(methods) - 1

	m.methodIndex = 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.methodIndex != last {
		t.Fatalf("left at index 0 should wrap to %d, got %d", last, m.methodIndex)
	}

	m.methodIndex = last
	updated2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated2.(Model)
	if m.methodIndex != 0 {
		t.Fatalf("right at last index should wrap to 0, got %d", m.methodIndex)
	}
}

func TestConcurrency_MinMaxEdges(t *testing.T) {
	m := NewModel()
	m.focus = focusConcurrency

	m.setConcurrency(1)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if got := m.concurrency(); got != 1 {
		t.Fatalf("left at min should stay 1, got %d", got)
	}

	m.setConcurrency(100)
	updated2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated2.(Model)
	if got := m.concurrency(); got != 100 {
		t.Fatalf("right at max should stay 100, got %d", got)
	}
}

func TestCtrlX_NotRunning(t *testing.T) {
	m := NewModel()
	m.status = "SYSTEM READY"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m = updated.(Model)

	if cmd != nil {
		t.Fatal("ctrl+x when not running should return nil cmd")
	}
	if m.status != "SYSTEM READY" {
		t.Fatalf("status should remain 'SYSTEM READY', got %q", m.status)
	}
}

func TestMouseMsg_NoCrash(t *testing.T) {
	m := NewModel()
	updated, cmd := m.Update(tea.MouseMsg{})
	_ = updated.(Model)
	if cmd != nil {
		t.Fatal("MouseMsg should return nil cmd")
	}
}

func TestStartRun_AlreadyRunning(t *testing.T) {
	m := NewModel()
	m.running = true
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)

	started, cmd := m.startRun()
	if cmd != nil {
		t.Fatal("startRun when already running should return nil cmd")
	}
	if !started.running {
		t.Fatal("should remain running")
	}
}

func TestCancelRun_NotRunning(t *testing.T) {
	m := NewModel()
	m.status = "IDLE"

	result := m.cancelRun()
	if result.status != "IDLE" {
		t.Fatalf("status should remain 'IDLE', got %q", result.status)
	}
}

func TestMultipleRunLifecycle(t *testing.T) {
	m := NewModel()
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
	}
	m.elapsed = 5 * time.Second

	started, _ := m.startRun()
	if !started.running {
		t.Fatal("should be running after startRun")
	}
	if len(started.results) != 0 {
		t.Fatal("results should be reset on startRun")
	}
	if started.elapsed != 0 {
		t.Fatal("elapsed should be reset on startRun")
	}

	cancelled := started.cancelRun()
	if cancelled.status != "CANCELLED" {
		t.Fatalf("status after cancel = %q, want 'CANCELLED'", cancelled.status)
	}
	if cancelled.running {
		t.Fatal("running should be false after cancelRun")
	}

	// cancelRun sets running=false and status=CANCELLED; runFinishedMsg
	// should not override the status when arriving afterwards.
	finished, _ := cancelled.Update(runFinishedMsg{})
	m2 := finished.(Model)
	if m2.running {
		t.Fatal("should not be running after runFinishedMsg")
	}
	if m2.status != "CANCELLED" {
		t.Fatalf("status after cancel+finish = %q, want 'CANCELLED'", m2.status)
	}

	restarted, cmd := m2.startRun()
	if cmd == nil {
		t.Fatal("startRun after cancel+finish should return a command")
	}
	if !restarted.running {
		t.Fatal("should be running after restart")
	}
}

func TestEsc_FromResultsNavigatesToURL(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.results = []model.Result{{Status: 200}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.focus != focusURL {
		t.Fatalf("esc from results should go to URL, got %v", m.focus)
	}
}

func TestEsc_FromInspectorWithResults(t *testing.T) {
	m := NewModel()
	m.inspector = true
	m.focus = focusResults
	m.results = []model.Result{
		{Status: 200},
	}
	m.selected = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.inspector {
		t.Fatal("esc should close inspector")
	}
	if m.focus != focusResults {
		t.Fatalf("focus should remain focusResults, got %v", m.focus)
	}
}

func TestResultMsg_NoEventChannel(t *testing.T) {
	m := NewModel()
	m.running = false
	m.eventCh = nil

	msg := resultMsg{Result: model.Result{Status: 200, Latency: 10 * time.Millisecond}}
	updated, cmd := m.Update(msg)
	m2 := updated.(Model)

	if cmd != nil {
		t.Fatal("resultMsg with no event channel should return nil cmd")
	}
	if len(m2.results) != 1 {
		t.Fatalf("result should be added, got %d results", len(m2.results))
	}
	if m2.summary.Total != 1 {
		t.Fatalf("summary should reflect 1 result, got %d", m2.summary.Total)
	}
}

func TestHeaderNavigation_UpDown(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow())

	if len(m.headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(m.headers))
	}

	m.selectedHead = 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selectedHead != 1 {
		t.Fatalf("down from index 0 should go to 1, got %d", m.selectedHead)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selectedHead != 0 {
		t.Fatalf("up from index 1 should go to 0, got %d", m.selectedHead)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selectedHead != 0 {
		t.Fatalf("up at index 0 should stay 0, got %d", m.selectedHead)
	}

	m.selectedHead = 1
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selectedHead != 1 {
		t.Fatalf("down at last index should stay at last, got %d", m.selectedHead)
	}
}

func TestHeaderNavigation_LeftRight(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders
	m.headers = append(m.headers, newHeaderRow())
	m.headerSubfocus = subfocusValue
	m, _ = m.syncFocus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m2 := updated.(Model)
	if m2.headerSubfocus != subfocusKey {
		t.Fatalf("left should set headerSubfocus to subfocusKey, got %d", m2.headerSubfocus)
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRight})
	m3 := updated.(Model)
	if m3.headerSubfocus != subfocusValue {
		t.Fatalf("right should set headerSubfocus to subfocusValue, got %d", m3.headerSubfocus)
	}
}

func TestHeaderNavigation_LeftRightAliases(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders
	m.headers = append(m.headers, newHeaderRow())
	m.headerSubfocus = subfocusValue
	m, _ = m.syncFocus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m2 := updated.(Model)
	if m2.headerSubfocus != subfocusKey {
		t.Fatalf("'h' should set headerSubfocus to subfocusKey, got %d", m2.headerSubfocus)
	}

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m3 := updated.(Model)
	if m3.headerSubfocus != subfocusValue {
		t.Fatalf("'l' should set headerSubfocus to subfocusValue, got %d", m3.headerSubfocus)
	}
}

func TestUpdateFocusedInput_Headers(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders
	m.headers = append(m.headers, newHeaderRow())
	m, _ = m.syncFocus()

	updated, _ := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m2 := updated.(Model)

	if m2.headers[0].Key.Value() != "a" {
		t.Fatal("header key input should contain typed character")
	}
}

func TestUpdateFocusedInput_Body(t *testing.T) {
	m := NewModel()
	m.focus = focusBody
	m, _ = m.syncFocus()

	updated, _ := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m2 := updated.(Model)

	if m2.bodyInput.Value() != "x" {
		t.Fatal("body input should contain typed character")
	}
}

func TestUpdateFocusedInput_URL(t *testing.T) {
	m := NewModel()
	m.focus = focusURL
	m, _ = m.syncFocus()

	updated, _ := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m2 := updated.(Model)

	if !strings.Contains(m2.urlInput.Value(), "/") {
		t.Fatal("URL input should contain typed character")
	}
}

func TestUpdateFocusedInput_CC(t *testing.T) {
	m := NewModel()
	m.focus = focusConcurrency
	m, _ = m.syncFocus()

	updated, _ := m.updateFocusedInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	m2 := updated.(Model)

	if !strings.Contains(m2.ccInput.Value(), "0") {
		t.Fatal("CC input should contain typed character")
	}
}

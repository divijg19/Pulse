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

func TestMethodSelection(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	// Tab advances from URL field to Payload domain (stays within request dialog)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.dialog != dialogRequest {
		t.Fatal("tab should not close request dialog")
	}
	if m.activeDomain != domainPayload {
		t.Fatal("tab from URL field should advance to payload domain")
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
	m.dialog = dialogRequest
	m.activeDomain = domainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey

	if m.dialog != dialogRequest {
		t.Fatal("request dialog should be active")
	}
	if m.activeDomain != domainPayload {
		t.Fatal("payload domain should be active")
	}
	if len(m.headers) != 0 {
		t.Fatalf("headers len = %d (expect 0, lazily initialized)", len(m.headers))
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

func TestViewSwitching(t *testing.T) {
	m := NewModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(Model)
	if m.view != viewLogs {
		t.Fatalf("view = %v (expected viewLogs)", m.view)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(Model)
	if m.view != viewTimeline {
		t.Fatalf("view = %v (expected viewTimeline)", m.view)
	}
}

func TestResultSelection_PageUpDown(t *testing.T) {
	m := NewModel()
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

func TestResultSelectionAndInspect(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: http.StatusOK},
		{Status: http.StatusInternalServerError},
	}

	// Navigate down to second result
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 1 {
		t.Fatalf("selected = %d", m.selected)
	}

	// Enter Inspect mode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.mode != modeInspect {
		t.Fatal("should enter inspect mode")
	}

	// ESC back to Observe
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.mode != modeObserve {
		t.Fatal("should return to observe mode")
	}
}

func TestHeaderAddRemove(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// ctrl+n adds a header row
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = updated.(Model)
	if len(m.headers) != 2 {
		t.Fatalf("after ctrl+n headers = %d (expected 2)", len(m.headers))
	}

	// ctrl+d removes last header row
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

func TestIsFollowingTail_Empty(t *testing.T) {
	m := NewModel()
	if !m.isFollowingTail() {
		t.Fatal("should follow tail when results are empty")
	}
}

func TestIsFollowingTail_AtBottom(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: 200},
		{Status: 200},
		{Status: 200},
	}
	m.selected = 2
	if !m.isFollowingTail() {
		t.Fatal("should follow tail when selected is last result")
	}
}

func TestIsFollowingTail_ScrolledUp(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: 200},
		{Status: 200},
		{Status: 200},
	}
	m.selected = 0
	if m.isFollowingTail() {
		t.Fatal("should not follow tail when scrolled away from bottom")
	}
}

func TestAutoScrollAdvancesSelection(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
	}
	m.selected = 0

	msg := resultMsg{Result: model.Result{Status: 200, Latency: 20 * time.Millisecond}}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if m.selected != 1 {
		t.Fatalf("after new result at bottom, selected = %d (expected 1)", m.selected)
	}
}

func TestAutoScrollDoesNotAdvanceWhenScrolledAway(t *testing.T) {
	m := NewModel()
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

func TestResultMsg_CapEnforced(t *testing.T) {
	m := NewModel()
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

func TestCtrlC_QuitIdle(t *testing.T) {
	m := NewModel()
	m.running = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("ctrl+c while idle should return a quit command")
	}
}

func TestQ_QuitIdle(t *testing.T) {
	m := NewModel()
	m.running = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("q while idle should return a quit command")
	}
}

func TestCtrlC_WhileRunning_ShowsConfirm(t *testing.T) {
	m := NewModel()
	m.running = true
	m.cancel = func() {}

	// First ctrl+c shows confirmation dialog
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)

	if cmd != nil {
		t.Fatal("first ctrl+c while running should return nil cmd (confirmation)")
	}
	if m.dialog != dialogConfirmQuit {
		t.Fatal("dialog should be confirmQuit after first ctrl+c")
	}

	// Second ctrl+c confirms quit
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("second ctrl+c should return a quit command")
	}
}

func TestConfirmQuit_OtherKeyCancels(t *testing.T) {
	m := NewModel()
	m.running = true
	m.dialog = dialogConfirmQuit
	m.cancel = func() {}

	// Any key other than q/ctrl+c/enter cancels
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(Model)

	if cmd != nil {
		t.Fatal("cancelling confirm should return nil cmd")
	}
	if m.dialog != dialogNone {
		t.Fatal("dialog should be none after cancel")
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

func TestObserve_EndpointDialogOpenClose(t *testing.T) {
	m := NewModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	if m.dialog != dialogRequest {
		t.Fatal("pressing e should open request dialog")
	}
	if m.activeDomain != domainRequest {
		t.Fatal("pressing e should set active domain to request")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.dialog != dialogNone {
		t.Fatal("pressing esc should close request dialog")
	}
}

func TestObserve_CCDialogOpenClose(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.ccInput.Focus()

	if m.dialog != dialogRequest {
		t.Fatal("request dialog should be active")
	}
	if m.activeDomain != domainExec {
		t.Fatal("execution domain should be active")
	}

	// Esc closes the request dialog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.dialog != dialogNone {
		t.Fatal("pressing esc should close request dialog")
	}
}

func TestObserve_PayloadDialogOpenClose(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey

	if m.dialog != dialogRequest {
		t.Fatal("request dialog should be active")
	}
	if m.activeDomain != domainPayload {
		t.Fatal("payload domain should be active")
	}
	if len(m.headers) != 0 {
		t.Fatal("headers should be lazily initialized")
	}

	// Esc closes the request dialog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.dialog != dialogNone {
		t.Fatal("pressing esc should close request dialog")
	}
}

func TestCCDialog_ArrowAdjustUp(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec

	initial := m.concurrency()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)

	if m.concurrency() != initial+1 {
		t.Fatalf("concurrency after up = %d (expected %d)", m.concurrency(), initial+1)
	}
}

func TestCCDialog_ArrowAdjustDown(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.setConcurrency(50)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	if m.concurrency() != 49 {
		t.Fatalf("concurrency after down = %d (expected 49)", m.concurrency())
	}
}

func TestEndpointDialog_EscCloses(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.dialog != dialogNone {
		t.Fatal("esc should close request dialog")
	}
}

func TestEndpointDialog_EnterDoesNotClose(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.dialog != dialogRequest {
		t.Fatal("enter should NOT close request dialog (esc-only)")
	}
}

func TestEndpointDialog_MethodSwitchingLeftRight(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldMethod

	methods := runconfig.AllowedMethods()

	// Start at index 0
	if m.methodIndex != 0 {
		t.Fatalf("methodIndex should start at 0, got %d", m.methodIndex)
	}

	// Right arrow increments
	updated, _ := m.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.methodIndex != 1 {
		t.Errorf("after right, methodIndex = %d, want 1", m.methodIndex)
	}

	// Left arrow decrements
	updated, _ = m.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.methodIndex != 0 {
		t.Errorf("after left, methodIndex = %d, want 0", m.methodIndex)
	}

	// Left at min does not go below 0
	updated, _ = m.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.methodIndex != 0 {
		t.Errorf("left at min should stay 0, got %d", m.methodIndex)
	}

	// Right at max does not exceed
	m.methodIndex = len(methods) - 1
	updated, _ = m.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.methodIndex != len(methods)-1 {
		t.Errorf("right at max should stay %d, got %d", len(methods)-1, m.methodIndex)
	}
}

func TestEndpointDialog_TabSwitchesField(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	// Tab should advance to payload domain (URL → Payload)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != domainPayload {
		t.Fatal("tab from URL field should advance to payload domain")
	}
}

func TestEndpointDialog_LeftRightOnUrlDoesNotChangeMethod(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	initialMethod := m.methodIndex

	// Left/right on URL should not change method
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.methodIndex != initialMethod {
		t.Fatal("left arrow on URL should not change method")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.methodIndex != initialMethod {
		t.Fatal("right arrow on URL should not change method")
	}
}

func TestCCDialog_EscCloses(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.ccInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.dialog != dialogNone {
		t.Fatal("esc should close request dialog")
	}
}

func TestCCDialog_EnterDoesNotClose(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.ccInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.dialog != dialogRequest {
		t.Fatal("enter should NOT close request dialog (esc-only)")
	}
}

func TestInspect_EnterFromObserve(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{{Status: 200}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.mode != modeInspect {
		t.Fatal("enter with results should enter inspect mode")
	}
}

func TestInspect_EscToObserve(t *testing.T) {
	m := NewModel()
	m.mode = modeInspect
	m.results = []model.Result{{Status: 200}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.mode != modeObserve {
		t.Fatal("esc from inspect should return to observe")
	}
}

func TestInspect_QuitOpensConfirm(t *testing.T) {
	m := NewModel()
	m.mode = modeInspect

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if m.dialog != dialogConfirmQuit {
		t.Fatal("q from inspect should open confirm dialog")
	}
}

func TestViewSwitch_PreservesSelection(t *testing.T) {
	m := NewModel()
	m.results = []model.Result{
		{Status: 200},
		{Status: 200},
	}
	m.selected = 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m2 := updated.(Model)
	if m2.selected != 1 {
		t.Fatalf("selection should be preserved after view switch, got %d", m2.selected)
	}
}

func TestEndpointDialog_NotClosableByWrongKey(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	// Typing text should go to urlInput, not close dialog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	if m.dialog != dialogRequest {
		t.Fatal("typing text should not close request dialog")
	}
	if !strings.Contains(m.urlInput.Value(), "/") {
		t.Fatal("URL input should contain typed character")
	}
}

func TestCCDialog_TypesDigits(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.ccInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	m = updated.(Model)

	if !strings.Contains(m.ccInput.Value(), "5") {
		t.Fatal("CC input should contain typed digit")
	}
}

func TestPayloadDialog_HeaderNavigationUpDown(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow())

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

func TestPayloadDialog_TabToBody(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Tab within payload: headers → body
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatalf("tab from headers should go to body, got selectedHead=%d", m.selectedHead)
	}

	// Tab from body → next domain (execution)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != domainExec {
		t.Fatalf("tab from body should advance to execution domain, got domain=%d", m.activeDomain)
	}
}

func TestStartRun_ResetsDialogAndMode(t *testing.T) {
	m := NewModel()
	m.mode = modeInspect
	m.dialog = dialogRequest
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)

	started, _ := m.startRun()
	if started.mode != modeObserve {
		t.Fatal("startRun should reset mode to Observe")
	}
	if started.dialog != dialogNone {
		t.Fatal("startRun should reset dialog to None")
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

func TestConfirmQuit_FromInspectMode(t *testing.T) {
	m := NewModel()
	m.mode = modeInspect
	m.running = true
	m.cancel = func() {}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if m.dialog != dialogConfirmQuit {
		t.Fatal("q from inspect should open confirm dialog")
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated.(Model)
	if cmd == nil {
		t.Fatal("confirm from inspect should return a quit command")
	}
}

func TestCCDialog_ClampAtMax(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.setConcurrency(100)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.concurrency() != 100 {
		t.Fatalf("concurrency at max should stay 100, got %d", m.concurrency())
	}
}

func TestCCDialog_ClampAtMin(t *testing.T) {
	m := NewModel()
	m.dialog = dialogRequest
	m.activeDomain = domainExec
	m.setConcurrency(1)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.concurrency() != 1 {
		t.Fatalf("concurrency at min should stay 1, got %d", m.concurrency())
	}
}

// Freezes the current Inspect navigation contract for v0.8.x discussion.
func TestInspectNavigationChangesSelection(t *testing.T) {
	m := NewModel()
	m.mode = modeInspect
	m.results = []model.Result{
		{Status: 200, Latency: 10 * time.Millisecond},
		{Status: 404, Latency: 20 * time.Millisecond},
		{Status: 500, Latency: 30 * time.Millisecond},
	}
	m.selected = 1

	// Up arrow should move selection up
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selected != 0 {
		t.Fatalf("up in inspect: selected = %d, want 0", m.selected)
	}

	// Down arrow should move selection down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 1 {
		t.Fatalf("down in inspect: selected = %d, want 1", m.selected)
	}

	// Down at last result should stay
	m.selected = 2
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 2 {
		t.Fatalf("down at last in inspect: selected = %d, want 2", m.selected)
	}
}

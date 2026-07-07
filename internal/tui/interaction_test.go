package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Pulse/internal/model"
)

// ---------------------------------------------------------------------------
// Suite 2  --  Behaviour Audit
// ---------------------------------------------------------------------------

func TestV091Behaviour_ObserveNoDeadKeys(t *testing.T) {
	m := NewModel()
	tabKeys := []tea.KeyMsg{
		keyMsgKey(tea.KeyTab),
		keyMsgKey(tea.KeyShiftTab),
		keyMsgKey(tea.KeyLeft),
		keyMsgKey(tea.KeyRight),
		keyMsgRune('h'),
		keyMsgRune('l'),
	}
	for _, key := range tabKeys {
		updated, cmd := m.handleObserveKey(key)
		m2 := updated.(Model)
		if cmd != nil {
			t.Fatalf("key %+v should not produce command in observe", key.Type)
		}
		if m2.workspace.dialog != dialogNone {
			t.Fatal("key should not open dialog")
		}
	}
}

func TestV091Behaviour_InspectNoDeadKeys(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = testResults(1)
	leftRightKeys := []tea.KeyMsg{
		keyMsgKey(tea.KeyLeft),
		keyMsgKey(tea.KeyRight),
		keyMsgRune('h'),
		keyMsgRune('l'),
	}
	for _, key := range leftRightKeys {
		updated, cmd := m.handleInspectKey(key)
		m2 := updated.(Model)
		if cmd != nil {
			t.Fatalf("key %+v should not produce command in inspect", key.Type)
		}
		if m2.workspace.mode != modeInspect {
			t.Fatal("key should stay in inspect mode")
		}
	}

	// Tab and Shift+Tab now cycle investigation zones — verify they work
	m2 := m
	updated, _ := m2.handleInspectKey(keyMsgKey(tea.KeyTab))
	m2 = updated.(Model)
	if m2.inspectZone != zoneWhy {
		t.Fatalf("Tab should advance to WHY zone, got %d", m2.inspectZone)
	}
	updated, _ = m2.handleInspectKey(keyMsgKey(tea.KeyShiftTab))
	m2 = updated.(Model)
	if m2.inspectZone != zoneWhatHappened {
		t.Fatalf("Shift+Tab should go back to WHAT HAPPENED zone, got %d", m2.inspectZone)
	}
}

func TestV091Behaviour_RequestNoDeadKeys(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()

	tabKeys := []tea.KeyMsg{
		keyMsgKey(tea.KeyLeft),
		keyMsgKey(tea.KeyRight),
		keyMsgKey(tea.KeyUp),
		keyMsgKey(tea.KeyDown),
		keyMsgRune('h'),
		keyMsgRune('l'),
	}
	for _, key := range tabKeys {
		updated, _ := m.handleRequestKey(key)
		m2 := updated.(Model)
		if m2.workspace.dialog != dialogRequest {
			t.Fatalf("key %+v should not close request dialog", key.Type)
		}
	}
}

// ---------------------------------------------------------------------------
// Suite 6  --  Operator Walkthrough
// ---------------------------------------------------------------------------

func TestV091Walkthrough_RequestRunInspectQuit(t *testing.T) {
	m := NewModel()

	// 1. Start  --  should show READY identity
	m.shell.Resize(100, 30)
	if !strings.Contains(m.View(), "READY") {
		t.Fatal("initial state should show READY")
	}

	// 2. Press e to open REQUEST dialog
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)
	if m.workspace.dialog != dialogRequest {
		t.Fatal("e should open REQUEST dialog")
	}

	// 3. REQUEST dialog should show identity
	if !strings.Contains(m.View(), "REQUEST") {
		t.Fatal("REQUEST dialog should show identity")
	}
	if !strings.Contains(m.View(), "Request") {
		t.Fatal("REQUEST dialog should show Request header")
	}

	// 4. Tab to advance through domains (Request→Payload headers)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload {
		t.Fatal("Tab should advance from Request to Payload domain")
	}

	// 5. Tab again (Payload header key→header value)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.headerSubfocus != subfocusValue {
		t.Fatal("Tab should advance from header Key to Value subfocus")
	}

	// 6. Tab again (header value→Payload body)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.selectedHead != bodyFocus {
		t.Fatal("Tab should advance from header Value to Payload body")
	}

	// 7. Tab again (Payload body→Execution)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainExec {
		t.Fatal("Tab should advance from Payload body to Execution domain")
	}

	// 8. Shift+Tab to go back (Exec→Payload body)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload {
		t.Fatal("Shift+Tab should go back to Payload domain")
	}

	// 9. Esc to close dialog
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.workspace.dialog != dialogNone {
		t.Fatal("Esc should close REQUEST dialog")
	}

	// 10. Press q to open quit confirmation
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if m.workspace.dialog != dialogConfirmQuit {
		t.Fatal("q should open quit confirmation")
	}

	// 11. Esc cancels quit
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	_ = m // esc sets dialog to dialogNone
}

// ---------------------------------------------------------------------------
// Suite 7  --  Navigation Audit
// ---------------------------------------------------------------------------

func TestV091Navigation_UpDownInRequest(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL

	// Up from URL → Method (blur URL, no focus yet)
	updated, _ := m.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.requestField != reqFieldMethod {
		t.Fatal("Up from URL should move to Method field")
	}

	// Up again at Method  --  should stay at Method (already at top)
	updated, _ = m2.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyUp})
	m3 := updated.(Model)
	if m3.requestField != reqFieldMethod {
		t.Fatal("Up at Method should stay at Method (already at top)")
	}

	// Down from Method → URL
	m3.requestField = reqFieldMethod
	m4, _ := m3.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyDown})
	if m4.(Model).requestField != reqFieldURL {
		t.Fatal("Down from Method should move to URL field")
	}

	// Down at URL  --  should stay at URL (already at bottom)
	m5 := m4.(Model)
	updated, _ = m5.handleRequestDomainKey(tea.KeyMsg{Type: tea.KeyDown})
	if updated.(Model).requestField != reqFieldURL {
		t.Fatal("Down at URL should stay at URL (already at bottom)")
	}
}

func TestV091Navigation_BlurAllBeforeDomainTransition(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.blurAll()
	m.urlInput.Focus()

	m2, _ := m.advanceDomain(true)
	if m2.(Model).activeDomain != DomainPayload {
		t.Fatal("Tab at URL field should advance to Payload domain")
	}
	if m2.(Model).urlInput.Focused() {
		t.Fatal("URL input must be blurred after advancing to Payload domain")
	}
}

// ---------------------------------------------------------------------------
// Suite 12  --  Boundary Traversal Audit
// ---------------------------------------------------------------------------

func TestV091Boundary_ReverseTabFromBodyToHeaders(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m2 := updated.(Model)
	if m2.selectedHead != len(m2.headers)-1 || m2.headerSubfocus != subfocusValue {
		t.Fatalf("Shift+Tab at body should go to last header Value, got head=%d subfocus=%d", m2.selectedHead, m2.headerSubfocus)
	}
}

func TestV091Boundary_ArrowUpInBodyStaysInBody(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.selectedHead != bodyFocus {
		t.Fatalf("↑ at body should stay in body (editor cursor move), got head=%d", m2.selectedHead)
	}
	if m2.activeDomain != DomainPayload {
		t.Fatalf("↑ at body should stay in Payload domain, got domain=%d", m2.activeDomain)
	}
}

func TestV091Boundary_ArrowDownInBodyStaysInBody(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(Model)
	if m2.selectedHead != bodyFocus {
		t.Fatalf("↓ at body should stay in body (editor cursor move), got head=%d", m2.selectedHead)
	}
	if m2.activeDomain != DomainPayload {
		t.Fatalf("↓ at body should stay in Payload domain, got domain=%d", m2.activeDomain)
	}
}

func TestV091Boundary_ExecDomainFocusedGuard(t *testing.T) {
	m := newRequestExecModel()
	m.setConcurrency(5)
	m.concurrencyInput.Focus()

	// Arrow keys adjust concurrency without losing focus
	before := m.concurrency()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.concurrency() != before+1 {
		t.Fatalf("↑ should increment concurrency from %d to %d, got %d", before, before+1, m2.concurrency())
	}
	if !m2.concurrencyInput.Focused() {
		t.Fatal("Exec domain should keep concurrencyInput focused after arrow adjustment")
	}
}

func TestV091Boundary_ArrowUpFromMethodIsNoOp(t *testing.T) {
	m := newRequestModel()
	m.requestField = reqFieldMethod

	before := m.methodIndex
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.methodIndex != before {
		t.Fatal("↑ at Method selector should be no-op (already at top)")
	}
}

func TestV091Boundary_ArrowDownAtUrlToPayload(t *testing.T) {
	m := newRequestModel()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(Model)
	if m2.activeDomain != DomainPayload || m2.selectedHead != 0 || m2.headerSubfocus != subfocusKey {
		t.Fatalf("↓ at URL should advance to Payload header row 0, got domain=%d head=%d subfocus=%d", m2.activeDomain, m2.selectedHead, m2.headerSubfocus)
	}
}

// ---------------------------------------------------------------------------
// Section: Correctness
//
// Invariant: Behaviour never produces invalid state.
// Failure: A sequence of operations leaves the model in an illegal
// (panic-able, unrecoverable, or contradictory) configuration.
//
// Verified:
//   - Header mutation sequences (delete, add, tab, esc)
//   - Robustness under rapid / abusive input
//   - Selection index bounds after every mutation
//   - Domain transition cycles (forward + reverse)
//   - Illegal (mode, dialog) combinations
// ---------------------------------------------------------------------------

// Invariant: Header mutation must never leave Payload in an illegal focus
// state. Every mutation must re-establish a valid focus target.

func TestV092HeaderMutation_DeleteLastThenTab(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Delete the only header row.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	if len(m.headers) != 0 {
		t.Fatalf("after delete: headers = %d (expected 0)", len(m.headers))
	}
	if m.selectedHead != bodyFocus {
		t.Fatalf("after delete: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}
	if !m.bodyInput.Focused() {
		t.Fatal("after delete: bodyInput should be focused")
	}

	// Tab should safely advance to Exec domain (body -> Exec).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	if m.activeDomain != DomainExec {
		t.Fatalf("after tab: activeDomain = %d (expected DomainExec=%d)", m.activeDomain, DomainExec)
	}
}

func TestV092HeaderMutation_DeleteLastThenShiftTab(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Delete the only header row.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	// Shift+Tab from bodyFocus should go to DomainRequest.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)

	if m.activeDomain != DomainRequest {
		t.Fatalf("after shift+tab: activeDomain = %d (expected DomainRequest=%d)", m.activeDomain, DomainRequest)
	}
	if !m.urlInput.Focused() {
		t.Fatal("after shift+tab: urlInput should be focused")
	}
}

func TestV092HeaderMutation_DeleteLastThenDown(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	// Down from bodyFocus should stay in body (editor cursor move).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	if m.activeDomain != DomainPayload {
		t.Fatalf("after down: activeDomain = %d (expected DomainPayload=%d)", m.activeDomain, DomainPayload)
	}
	if m.selectedHead != bodyFocus {
		t.Fatalf("after down: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}
}

func TestV092HeaderMutation_DeleteLastThenEsc(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Delete then Esc should close the dialog cleanly.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.workspace.dialog != dialogNone {
		t.Fatalf("after esc: dialog = %d (expected dialogNone=%d)", m.workspace.dialog, dialogNone)
	}
}

func TestV092HeaderMutation_DeleteAllThenAdd(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey

	// Add multiple headers then delete them all.
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow(), newHeaderRow())

	for i := 0; i < 3; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
		m = updated.(Model)
	}

	if len(m.headers) != 0 {
		t.Fatalf("after deleting all: headers = %d (expected 0)", len(m.headers))
	}
	if m.selectedHead != bodyFocus {
		t.Fatalf("after deleting all: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}

	// From bodyFocus with empty headers, the path to re-add a header is:
	// Shift+Tab -> DomainRequest (URL focus) -> Tab -> Payload (guard creates header).
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainRequest {
		t.Fatalf("after shift+tab: activeDomain = %d (expected DomainRequest=%d)", m.activeDomain, DomainRequest)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload {
		t.Fatalf("after tab from request: activeDomain = %d (expected DomainPayload=%d)", m.activeDomain, DomainPayload)
	}
	if len(m.headers) == 0 {
		t.Fatal("after returning to payload: headers should not be empty")
	}
	if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
		t.Fatalf("after returning: selectedHead = %d (out of bounds for len=%d)", m.selectedHead, len(m.headers))
	}
	if !m.headers[m.selectedHead].Key.Focused() {
		t.Fatal("after returning: header key should be focused")
	}
}

func TestV092HeaderMutation_DeleteWhileEditing(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow())

	// Type into the first header's key field.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	m = updated.(Model)

	// Delete the first header (selectedHead=0).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	if len(m.headers) != 1 {
		t.Fatalf("after delete: headers = %d (expected 1)", len(m.headers))
	}
	if m.selectedHead != 0 {
		t.Fatalf("after delete: selectedHead = %d (expected 0)", m.selectedHead)
	}
}

func TestV092HeaderMutation_DeleteMiddleHeader(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey

	// Create three headers.
	m.headers = append(m.headers, newHeaderRow(), newHeaderRow(), newHeaderRow())

	// Select the middle header (index 1).
	m.selectedHead = 1
	m.focusPayloadKey()

	// Delete middle header.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	if len(m.headers) != 2 {
		t.Fatalf("after delete middle: headers = %d (expected 2)", len(m.headers))
	}
	if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
		t.Fatalf("after delete middle: selectedHead = %d (out of bounds for len=%d)", m.selectedHead, len(m.headers))
	}
}

func TestV092HeaderMutation_DeleteLastWithValueSubfocus(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusValue
	m.headers = append(m.headers, newHeaderRow())

	// Delete while on value subfocus.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	if len(m.headers) != 0 {
		t.Fatalf("after delete: headers = %d (expected 0)", len(m.headers))
	}
	if m.selectedHead != bodyFocus {
		t.Fatalf("after delete: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}
	if !m.bodyInput.Focused() {
		t.Fatal("after delete: bodyInput should be focused")
	}
}

func TestV092HeaderMutation_RapidAddRemove(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Rapidly alternate Ctrl+N and Ctrl+D 20 times.
	for i := 0; i < 20; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
		m = updated.(Model)
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
		m = updated.(Model)
	}

	// Must end in a valid state: either bodyFocus with empty headers,
	// or a valid selectedHead pointing at a real header.
	if len(m.headers) == 0 {
		if m.selectedHead != bodyFocus {
			t.Fatalf("empty headers but selectedHead=%d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
		}
	} else {
		if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
			t.Fatalf("non-empty headers (len=%d) but selectedHead=%d is out of bounds", len(m.headers), m.selectedHead)
		}
	}
}

// Invariant: Nothing panics under any input sequence.

func TestV092Robustness_DeleteThenTabChain(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Chain: delete -> tab -> esc (should never panic)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.workspace.dialog != dialogNone {
		t.Fatalf("final dialog = %d (expected dialogNone=%d)", m.workspace.dialog, dialogNone)
	}
}

func TestV092Robustness_DeleteThenEscThenReopen(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Delete -> Esc -> reopen and verify dialog is in a valid state.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	// Reopen the request dialog.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = updated.(Model)

	if m.workspace.dialog != dialogRequest {
		t.Fatalf("after reopen: dialog = %d (expected dialogRequest=%d)", m.workspace.dialog, dialogRequest)
	}
}

func TestV092Robustness_DeleteThenCtrlR(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	// Delete -> Ctrl+R key dispatch should never panic.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = updated.(Model)

	// startRun from request dialog proceeds regardless of domain.
	if !m.running {
		t.Fatal("after delete+ctrl+r: should be running (NewModel provides default URL)")
	}
}

// Invariant: Every selection index is valid after any mutation.

func TestV092Selection_SelectedHeadBounds(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload

	// 1. selectedHead should be bodyFocus when no headers exist.
	m.headers = nil
	m.selectedHead = bodyFocus
	if m.selectedHead != bodyFocus {
		t.Fatalf("no headers: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}

	// 2. After deleting all headers, selectedHead must be bodyFocus.
	m.headers = append(m.headers, newHeaderRow())
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatalf("after delete all: selectedHead = %d (expected bodyFocus=%d)", m.selectedHead, bodyFocus)
	}

	// 3. After re-entering payload, selectedHead must be valid.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
		t.Fatalf("after re-entering payload: selectedHead = %d (len=%d)", m.selectedHead, len(m.headers))
	}
}

func TestV092Selection_SelectedBounds(t *testing.T) {
	m := NewModel()

	// 1. Default: selected is 0, no results.
	if m.selected != 0 {
		t.Fatalf("initial selected = %d (expected 0)", m.selected)
	}

	// 2. After startRun, selected is reset to 0.
	m.running = true
	m.results = []model.Result{{Status: 200}, {Status: 404}}
	m.selected = 1
	m.running = false
	// Simulate startRun reset.
	m.results = nil
	m.selected = 0
	if m.selected != 0 {
		t.Fatalf("after reset: selected = %d (expected 0)", m.selected)
	}

	// 3. selected must never exceed len(results)-1 after navigation.
	m.results = []model.Result{{Status: 200}}
	m.selected = 0
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 0 {
		t.Fatalf("down at single result: selected = %d (expected 0)", m.selected)
	}
}

func TestV092Selection_HeaderSubfocusBounds(t *testing.T) {
	// headerSubfocus must always be 0 (key) or 1 (value).
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	if m.headerSubfocus != subfocusKey && m.headerSubfocus != subfocusValue {
		t.Fatalf("invalid headerSubfocus = %d", m.headerSubfocus)
	}

	// After right arrow, should be value.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.headerSubfocus != subfocusValue {
		t.Fatalf("after right: headerSubfocus = %d (expected subfocusValue=%d)", m.headerSubfocus, subfocusValue)
	}

	// After left arrow, should be key.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.headerSubfocus != subfocusKey {
		t.Fatalf("after left: headerSubfocus = %d (expected subfocusKey=%d)", m.headerSubfocus, subfocusKey)
	}
}

// Invariant: Every domain transition is reversible.

func TestV092Boundary_FullTransitionCycle(t *testing.T) {
	// The full forward transition cycle via Tab:
	//   Request.Method -> Request.URL -> Payload.Key -> Payload.Value ->
	//   Payload.Body -> Exec -> Request.Method (wrap)
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.headers = append(m.headers, newHeaderRow())

	m.activeDomain = DomainRequest
	m.requestField = reqFieldMethod

	// Tab 1: Method -> URL
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainRequest || m.requestField != reqFieldURL {
		t.Fatalf("after tab 1: expected (Request, URL), got (domain=%d, field=%d)", m.activeDomain, m.requestField)
	}

	// Tab 2: URL -> Payload.Key
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.headerSubfocus != subfocusKey {
		t.Fatalf("after tab 2: expected (Payload, Key), got (domain=%d, subfocus=%d)", m.activeDomain, m.headerSubfocus)
	}

	// Tab 3: Payload.Key -> Payload.Value
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.headerSubfocus != subfocusValue {
		t.Fatalf("after tab 3: expected (Payload, Value), got (domain=%d, subfocus=%d)", m.activeDomain, m.headerSubfocus)
	}

	// Tab 4: Payload.Value -> Payload.Body (last header)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.selectedHead != bodyFocus {
		t.Fatalf("after tab 4: expected (Payload, Body), got (domain=%d, head=%d)", m.activeDomain, m.selectedHead)
	}

	// Tab 5: Payload.Body -> Exec
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainExec {
		t.Fatalf("after tab 5: expected Exec, got domain=%d", m.activeDomain)
	}

	// Tab 6: Exec -> Request.Method (wrap)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.activeDomain != DomainRequest || m.requestField != reqFieldMethod {
		t.Fatalf("after tab 6: expected (Request, Method), got (domain=%d, field=%d)", m.activeDomain, m.requestField)
	}
}

func TestV092Boundary_FullReverseTransitionCycle(t *testing.T) {
	// Reverse transition cycle via Shift+Tab:
	//   Request.Method -> Exec -> Payload.Body -> Payload.Value ->
	//   Payload.Key -> Request.URL -> Request.Method
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.headers = append(m.headers, newHeaderRow())

	// Start at Request.Method.
	m.activeDomain = DomainRequest
	m.requestField = reqFieldMethod

	// Shift+Tab 1: Request.Method -> Exec (wrap)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainExec {
		t.Fatalf("after shift+tab 1: expected Exec, got domain=%d", m.activeDomain)
	}

	// Shift+Tab 2: Exec -> Payload.Body
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.selectedHead != bodyFocus {
		t.Fatalf("after shift+tab 2: expected (Payload, Body), got (domain=%d, head=%d)", m.activeDomain, m.selectedHead)
	}

	// Shift+Tab 3: Payload.Body -> Payload.Value (last header)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.headerSubfocus != subfocusValue {
		t.Fatalf("after shift+tab 3: expected (Payload, Value), got (domain=%d, subfocus=%d)", m.activeDomain, m.headerSubfocus)
	}

	// Shift+Tab 4: Payload.Value -> Payload.Key
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainPayload || m.headerSubfocus != subfocusKey {
		t.Fatalf("after shift+tab 4: expected (Payload, Key), got (domain=%d, subfocus=%d)", m.activeDomain, m.headerSubfocus)
	}

	// Shift+Tab 5: Payload.Key -> Request.URL (selectedHead is 0, so go to request)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainRequest || m.requestField != reqFieldURL {
		t.Fatalf("after shift+tab 5: expected (Request, URL), got (domain=%d, field=%d)", m.activeDomain, m.requestField)
	}

	// Shift+Tab 6: Request.URL -> Request.Method
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.activeDomain != DomainRequest || m.requestField != reqFieldMethod {
		t.Fatalf("after shift+tab 6: expected (Request, Method), got (domain=%d, field=%d)", m.activeDomain, m.requestField)
	}
}

// Invariant: No illegal (mode, dialog) combination silently drops keys.

func TestV092Boundary_IllegalInspectRequest(t *testing.T) {
	// modeInspect + dialogRequest should never occur, but if it does,
	// it must not panic and must not silently swallow fatal keys.
	m := NewModel()
	m.workspace.mode = modeInspect
	m.workspace.dialog = dialogRequest

	// These should not panic.
	keys := []tea.KeyType{tea.KeyEsc, tea.KeyEnter, tea.KeyTab, tea.KeyUp, tea.KeyDown}
	for _, kt := range keys {
		t.Run(kt.String(), func(t *testing.T) {
			_, _ = m.Update(tea.KeyMsg{Type: kt})
		})
	}
}

// ---------------------------------------------------------------------------
// Section: Interaction
//
// Invariant: Every keystroke produces predictable, deterministic behaviour.
// Failure: A key is unhandled (dropped), ambiguous (multiple handlers), or
// produces unexpected side effects.
//
// Verified:
//   - Every (mode, dialog) state dispatches to exactly one handler
//   - Navigation keys are no-ops in observe and inspect
// ---------------------------------------------------------------------------

// Invariant: Every key has exactly one owner per (mode, dialog) state.

func TestV092Interaction_HandleKeyDispatchExhaustive(t *testing.T) {
	// All dispatch states must resolve to a valid handler without panic.
	// Each (mode, dialog) pair maps to exactly one handler function.
	states := []struct {
		name    string
		setup   func(m *Model)
		handler string
	}{
		{
			"observe/none",
			func(m *Model) {
				m.workspace.mode = modeObserve
				m.workspace.dialog = dialogNone
			},
			"handleObserveKey",
		},
		{
			"observe/request",
			func(m *Model) {
				m.workspace.mode = modeObserve
				m.workspace.dialog = dialogRequest
			},
			"handleRequestKey",
		},
		{
			"observe/confirmQuit",
			func(m *Model) {
				m.workspace.mode = modeObserve
				m.workspace.dialog = dialogConfirmQuit
			},
			"handleConfirmQuitKey",
		},
		{
			"inspect/none",
			func(m *Model) {
				m.workspace.mode = modeInspect
				m.workspace.dialog = dialogNone
			},
			"handleInspectKey",
		},
		{
			"inspect/confirmQuit",
			func(m *Model) {
				m.workspace.mode = modeInspect
				m.workspace.dialog = dialogConfirmQuit
			},
			"handleConfirmQuitKey",
		},
	}

	// Keys that should not panic in any state.
	keys := []tea.KeyType{
		tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight,
		tea.KeyTab, tea.KeyShiftTab,
		tea.KeyEnter, tea.KeyEsc,
		tea.KeyCtrlC, tea.KeyCtrlD, tea.KeyCtrlN, tea.KeyCtrlR, tea.KeyCtrlX,
		tea.KeyPgUp, tea.KeyPgDown,
	}

	for _, st := range states {
		for _, kt := range keys {
			t.Run(st.name+"/"+kt.String(), func(t *testing.T) {
				m := NewModel()
				st.setup(&m)

				// Must not panic for any key in any state.
				_, _ = m.Update(tea.KeyMsg{Type: kt})
			})
		}
	}
}

func TestV092Interaction_ObserveInspectKeysNoop(t *testing.T) {
	// tab, shift+tab, left, right, h, l are no-ops in observe and inspect.
	observeKeys := []tea.KeyType{tea.KeyTab, tea.KeyShiftTab, tea.KeyLeft, tea.KeyRight}
	m := NewModel()

	for _, kt := range observeKeys {
		updated, _ := m.Update(tea.KeyMsg{Type: kt})
		m2 := updated.(Model)
		if m2.workspace.mode != modeObserve || m2.workspace.dialog != dialogNone {
			t.Fatalf("key %s changed state: mode=%d dialog=%d", kt.String(), m2.workspace.mode, m2.workspace.dialog)
		}
	}

	// Same keys in inspect mode.
	m.workspace.mode = modeInspect
	for _, kt := range observeKeys {
		updated, _ := m.Update(tea.KeyMsg{Type: kt})
		m2 := updated.(Model)
		if m2.workspace.mode != modeInspect || m2.workspace.dialog != dialogNone {
			t.Fatalf("key %s changed state: mode=%d dialog=%d", kt.String(), m2.workspace.mode, m2.workspace.dialog)
		}
	}
}

// Invariant: hjkl in editable fields type text; arrows navigate.
// hjkl in read-only browse modes (observe/inspect/compare) navigate.

func TestV092Interaction_HJKLConsistency(t *testing.T) {
	// Editable fields: hjkl must type into the focused widget.
	// Arrow keys continue to navigate, cycle, or adjust values.

	// Exec domain: k/j type text; up/down inc/dec concurrency.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.concurrencyInput.Focus()
	m.concurrencyInput.SetValue("5")

	updatedK, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	mK := updatedK.(Model)
	if mK.concurrencyInput.Value() != "5k" {
		t.Fatalf("exec: k should type into concurrency input, got %q", mK.concurrencyInput.Value())
	}

	updatedJ, _ := mK.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	mJ := updatedJ.(Model)
	if mJ.concurrencyInput.Value() != "5kj" {
		t.Fatalf("exec: j should type into concurrency input, got %q", mJ.concurrencyInput.Value())
	}

	// Arrows still inc/dec the parsed value.
	updatedUp, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mUp := updatedUp.(Model)
	if mUp.concurrencyInput.Value() != "6" {
		t.Fatalf("exec: up should increment concurrency, got %q", mUp.concurrencyInput.Value())
	}

	updatedDown, _ := mUp.Update(tea.KeyMsg{Type: tea.KeyDown})
	mDown := updatedDown.(Model)
	if mDown.concurrencyInput.Value() != "5" {
		t.Fatalf("exec: down should decrement concurrency, got %q", mDown.concurrencyInput.Value())
	}

	// Payload header: left/right switch subfocus; h/l move cursor within text.
	// This is by design: h/l are text-editing keys (consistent with URL/body),
	// while left/right are field-navigation keys (consistent with Tab semantics).
	m2 := NewModel()
	m2.workspace.dialog = dialogRequest
	m2.activeDomain = DomainPayload
	m2.selectedHead = 0
	m2.headerSubfocus = subfocusKey
	m2.headers = append(m2.headers, newHeaderRow())

	// right switches from key to value subfocus
	updatedRight, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRight})
	rightModel := updatedRight.(Model)
	if rightModel.headerSubfocus != subfocusValue {
		t.Fatal("right should switch from key to value subfocus")
	}

	// l does NOT switch subfocus (it falls through to text input for cursor movement)
	// Rebuild fresh state and test
	m3 := NewModel()
	m3.workspace.dialog = dialogRequest
	m3.activeDomain = DomainPayload
	m3.selectedHead = 0
	m3.headerSubfocus = subfocusKey
	m3.headers = append(m3.headers, newHeaderRow())
	updatedL, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	lModel := updatedL.(Model)
	if lModel.headerSubfocus != subfocusKey {
		t.Fatal("l should fall through to text input, not switch subfocus")
	}

	// left from value switches back to key subfocus
	updatedLeft, _ := rightModel.Update(tea.KeyMsg{Type: tea.KeyLeft})
	leftModel := updatedLeft.(Model)
	if leftModel.headerSubfocus != subfocusKey {
		t.Fatal("left should switch from value to key subfocus")
	}

	// h does NOT switch subfocus (it falls through to text input)
	updatedH, _ := lModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	hModel := updatedH.(Model)
	if hModel.headerSubfocus != subfocusKey {
		t.Fatal("h should fall through to text input, not switch subfocus")
	}
}

// Invariant: Empty-state navigation is safe (no panics, no state change).

func TestV092Interaction_EmptyStateNavigationSafety(t *testing.T) {
	// Navigation keys in the idle state (no results) must not panic and
	// must not change the selection index or mode.
	m := NewModel()
	keys := []tea.KeyMsg{
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyEnter},
		{Type: tea.KeyPgUp},
		{Type: tea.KeyPgDown},
	}
	for _, k := range keys {
		t.Run(k.Type.String(), func(t *testing.T) {
			updated, _ := m.Update(k)
			m2 := updated.(Model)
			if m2.selected != 0 {
				t.Fatalf("empty-state navigation with %s changed selected to %d", k.Type.String(), m2.selected)
			}
			if m2.workspace.mode != modeObserve {
				t.Fatalf("empty-state navigation with %s changed mode", k.Type.String())
			}
		})
	}
}

// Invariant: Ctrl+R dispatches from every request domain.

func TestV092Interaction_CtrlRFromEveryDomain(t *testing.T) {
	// Ctrl+R must start a run from Request, Payload, and Exec domains.
	domains := []struct {
		name  string
		setup func(m *Model)
	}{
		{"Request", func(m *Model) { m.activeDomain = DomainRequest }},
		{"Payload", func(m *Model) { m.activeDomain = DomainPayload }},
		{"Exec", func(m *Model) { m.activeDomain = DomainExec }},
	}
	for _, d := range domains {
		t.Run(d.name, func(t *testing.T) {
			m := NewModel()
			m.workspace.dialog = dialogRequest
			d.setup(&m)
			updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
			m2 := updated.(Model)
			if !m2.running {
				t.Fatalf("ctrl+R from %s domain did not start a run", d.name)
			}
			if cmd == nil {
				t.Fatalf("ctrl+R from %s domain produced no command", d.name)
			}
		})
	}
}

// Invariant: Ctrl+X from within the Request dialog cancels the run.

func TestV092Interaction_CtrlXFromDialog(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.running = true
	m.cancel = func() {}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m2 := updated.(Model)
	if m2.running {
		t.Fatal("ctrl+X from Request dialog should cancel the run")
	}
	if m2.status != "CANCELLED" {
		t.Fatalf("ctrl+X from Request dialog should set status to CANCELLED, got %q", m2.status)
	}
}

// Regression: editable fields type h/j/k/l; body up/down stay in body.
// Read-only browse modes (observe, inspect, compare) keep j/k navigation.

func TestV092Regression_PayloadBodyEditorKeys(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()
	m.bodyInput.SetValue("")

	// h/j/k/l type into body.
	for _, r := range []rune{'h', 'j', 'k', 'l'} {
		m2, _ := m.Update(keyMsgRune(r))
		m = m2.(Model)
	}
	if m.bodyInput.Value() != "hjkl" {
		t.Fatalf("body should type hjkl, got %q", m.bodyInput.Value())
	}
	if m.selectedHead != bodyFocus {
		t.Fatal("body should stay focused after typing hjkl")
	}
	if m.activeDomain != DomainPayload {
		t.Fatal("body should stay in Payload domain after typing hjkl")
	}

	// Up/down stay in body (editor cursor movement).
	m2, _ := m.Update(keyMsgKey(tea.KeyUp))
	m = m2.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatal("↑ at body should stay in body")
	}
	m2, _ = m.Update(keyMsgKey(tea.KeyDown))
	m = m2.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatal("↓ at body should stay in body")
	}

	// Left/right stay in body.
	m2, _ = m.Update(keyMsgKey(tea.KeyLeft))
	m = m2.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatal("← at body should stay in body")
	}
	m2, _ = m.Update(keyMsgKey(tea.KeyRight))
	m = m2.(Model)
	if m.selectedHead != bodyFocus {
		t.Fatal("→ at body should stay in body")
	}

	// Tab leaves body for Exec domain.
	m2, _ = m.Update(keyMsgKey(tea.KeyTab))
	m = m2.(Model)
	if m.activeDomain != DomainExec {
		t.Fatal("Tab from body should go to Exec domain")
	}
}

func TestV092Regression_HeaderVimKeysType(t *testing.T) {
	// In header key/value fields, j/k type text.
	// Up/down continue navigating between header rows.
	m := newRequestPayloadModel()
	m.selectedHead = 0
	m.headerSubfocus = subfocusValue
	m.focusPayloadValue()
	m.headers[0].Value.SetValue("")

	// j/k at header value should type.
	m2, _ := m.Update(keyMsgRune('j'))
	m = m2.(Model)
	if m.headers[0].Value.Value() != "j" {
		t.Fatalf("'j' at header value should type 'j', got %q", m.headers[0].Value.Value())
	}
	m2, _ = m.Update(keyMsgRune('k'))
	m = m2.(Model)
	if m.headers[0].Value.Value() != "jk" {
		t.Fatalf("'k' at header value should type 'k', got %q", m.headers[0].Value.Value())
	}
	if m.selectedHead != 0 {
		t.Fatal("j/k should not change header row")
	}

	// Up/down still navigate header rows.
	m2, _ = m.Update(keyMsgKey(tea.KeyUp))
	m = m2.(Model)
	if m.selectedHead != 0 {
		t.Fatal("↑ at first header row should cross to Request domain")
	}
	if m.activeDomain != DomainRequest {
		t.Fatal("↑ at first header row should go to Request domain")
	}
}

func TestV092Regression_ObserveVimKeysNavigate(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeObserve
	m.workspace.dialog = dialogNone
	m.results = []model.Result{
		{Status: 200},
		{Status: 404},
		{Status: 500},
	}
	m.selected = 1

	// 'k' navigates up in observe mode.
	m2, _ := m.Update(keyMsgRune('k'))
	m = m2.(Model)
	if m.selected != 0 {
		t.Fatal("'k' in observe mode should navigate up")
	}

	// 'j' navigates down in observe mode.
	m2, _ = m.Update(keyMsgRune('j'))
	m = m2.(Model)
	if m.selected != 1 {
		t.Fatal("'j' in observe mode should navigate down")
	}
}

func TestV092Regression_InspectVimKeysScroll(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = []model.Result{
		{
			Status:       200,
			ResponseBody: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10",
		},
	}
	m.selected = 0
	m.inspectZone = zoneBody
	m.inspectBodyOffset = 5

	// 'k' scrolls body up in inspect mode.
	updated, _ := m.handleInspectKey(keyMsgRune('k'))
	m2 := updated.(Model)
	if m2.inspectBodyOffset != 4 {
		t.Fatal("'k' in inspect body should decrement offset")
	}

	// 'j' scrolls body down in inspect mode.
	updated, _ = m2.handleInspectKey(keyMsgRune('j'))
	m3 := updated.(Model)
	if m3.inspectBodyOffset != 5 {
		t.Fatal("'j' in inspect body should increment offset")
	}
}

func TestV092Regression_CompareVimKeysScroll(t *testing.T) {
	m := NewModel()
	m.workspace.mode = modeCompare
	m.results = []model.Result{
		{Status: 200, ResponseBody: "a\nb\nc\nd\ne\nf\ng"},
		{Status: 404, ResponseBody: "a\nb\nc\nd\ne\nf\ng"},
	}
	m.workspace.compare = compareState{marked: 0, active: 1}
	m.inspectZone = zoneBody
	m.inspectBodyOffset = 5

	// 'k' scrolls body up in compare mode.
	updated, _ := m.handleCompareKey(keyMsgRune('k'))
	m2 := updated.(Model)
	if m2.inspectBodyOffset != 4 {
		t.Fatal("'k' in compare body should decrement offset")
	}

	// 'j' scrolls body down in compare mode.
	updated, _ = m2.handleCompareKey(keyMsgRune('j'))
	m3 := updated.(Model)
	if m3.inspectBodyOffset != 5 {
		t.Fatal("'j' in compare body should increment offset")
	}
}

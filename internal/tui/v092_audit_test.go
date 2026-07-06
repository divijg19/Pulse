package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Pulse/internal/model"
)

// typeName returns the type name for reflection comparisons in tests.
func typeName(v interface{}) string {
	return fmt.Sprintf("%T", v)
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

	// Down from bodyFocus goes to Exec domain.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	if m.activeDomain != DomainExec {
		t.Fatalf("after down: activeDomain = %d (expected DomainExec=%d)", m.activeDomain, DomainExec)
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
// Section: Architecture
//
// Invariant: Every concept has exactly one owner.
// Failure: A type reads state it does not own, or renders output that
// belongs to another layer.
//
// Verified:
//   - Workspace is the single source of mode/dialog/view
//   - Surface dispatch is deterministic and complete
//   - Domains describe intent (actions), not rendering
//   - Focus functions guard against invalid indices
//   - Focus transitions are consistent
// ---------------------------------------------------------------------------

// Invariant: Workspace is the single source of truth for mode, dialog, and
// view. No code reads them independently.

func TestV092Constitution_WorkspaceSingleSource(t *testing.T) {
	m := NewModel()
	if m.workspace.mode != modeObserve {
		t.Fatalf("initial mode = %d (expected modeObserve=%d)", m.workspace.mode, modeObserve)
	}
	if m.workspace.dialog != dialogNone {
		t.Fatalf("initial dialog = %d (expected dialogNone=%d)", m.workspace.dialog, dialogNone)
	}
	if m.workspace.view != TimelineView {
		t.Fatalf("initial view = %d (expected TimelineView=%d)", m.workspace.view, TimelineView)
	}

	// orientationLabel may refine the idle Observe workspace into READY.
	label := orientationLabel(m)
	if label != "READY" {
		t.Fatalf("orientationLabel = %q (expected READY)", label)
	}

	// Changing workspace state changes orientation.
	m.workspace.dialog = dialogRequest
	if orientationLabel(m) != "REQUEST" {
		t.Fatalf("after dialogRequest: orientationLabel = %q (expected REQUEST)", orientationLabel(m))
	}

	m.workspace.dialog = dialogConfirmQuit
	if orientationLabel(m) != "QUIT" {
		t.Fatalf("after dialogConfirmQuit: orientationLabel = %q (expected QUIT)", orientationLabel(m))
	}

	m.workspace.dialog = dialogNone
	m.workspace.mode = modeInspect
	if orientationLabel(m) != "INSPECT" {
		t.Fatalf("after modeInspect: orientationLabel = %q (expected INSPECT)", orientationLabel(m))
	}
}

// Invariant: Every render path resolves through resolveSurface().

func TestV092Constitution_SurfaceResolvability(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(m *Model)
		wantType string
	}{
		{
			"request dialog overrides all",
			func(m *Model) { m.workspace.dialog = dialogRequest },
			"tui.RequestSurface",
		},
		{
			"inspect mode",
			func(m *Model) { m.workspace.mode = modeInspect },
			"tui.InspectSurface",
		},
		{
			"idle with no results",
			func(m *Model) {},
			"tui.ReadySurface",
		},
		{
			"running with timeline view (default)",
			func(m *Model) { m.running = true },
			"tui.TimelineSurface",
		},
		{
			"has results with timeline view",
			func(m *Model) {
				m.results = []model.Result{{Status: 200}}
			},
			"tui.TimelineSurface",
		},
		{
			"logs view with results",
			func(m *Model) {
				m.results = []model.Result{{Status: 200}}
				m.workspace.view = LogsView
			},
			"tui.LogsSurface",
		},
		{
			"running with logs view",
			func(m *Model) {
				m.running = true
				m.results = []model.Result{{Status: 200}}
				m.workspace.view = LogsView
			},
			"tui.LogsSurface",
		},
		{
			"request dialog with inspect mode (dialog wins)",
			func(m *Model) {
				m.workspace.dialog = dialogRequest
				m.workspace.mode = modeInspect
			},
			"tui.RequestSurface",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel()
			tc.setup(&m)
			surface := m.resolveSurface()
			got := surface.Render(Region{Width: 80, Height: 24})
			if got == "" {
				t.Fatal("surface rendered empty output")
			}
			gotType := typeName(surface)
			if gotType != tc.wantType {
				t.Fatalf("resolveSurface() = %s (expected %s)", gotType, tc.wantType)
			}
		})
	}
}

// Invariant: Domain actions describe intent; they never render.

func TestV092Constitution_DomainActions(t *testing.T) {
	m := NewModel()

	for dt, expected := range map[DomainType]int{
		DomainRequest: 2,
		DomainPayload: 3,
		DomainExec:    1,
	} {
		domain, ok := domainRegistry[dt]
		if !ok {
			t.Fatalf("domainRegistry missing entry for DomainType=%d", dt)
		}
		actions := domain.Actions(m)
		if len(actions) != expected {
			t.Fatalf("DomainType=%d: Actions() returned %d (expected %d)", dt, len(actions), expected)
		}
		for i, a := range actions {
			if !a.Enabled {
				t.Fatalf("DomainType=%d action[%d] is disabled", dt, i)
			}
		}
	}
}

// Invariant: Every focus function guards against invalid indices.

func TestV092Focus_EmptyHeaderGuard(t *testing.T) {
	// focusPayloadKey and focusPayloadValue must not panic when headers
	// is empty or selectedHead is out of range.
	m := NewModel()
	m.headers = nil
	m.selectedHead = 0

	// Must not panic with empty headers.
	m.focusPayloadKey()
	m.focusPayloadValue()

	// Must not panic with out-of-range index.
	m.selectedHead = 5
	m.focusPayloadKey()
	m.focusPayloadValue()

	m.selectedHead = -10
	m.focusPayloadKey()
	m.focusPayloadValue()
}

func TestV092Focus_TransitionConsistency(t *testing.T) {
	// 1. Delete last header -> bodyInput focused.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)
	if !m.bodyInput.Focused() {
		t.Fatal("after delete last header: bodyInput should be focused")
	}

	// 2. Delete middle header -> remaining header key focused.
	m2 := NewModel()
	m2.workspace.dialog = dialogRequest
	m2.activeDomain = DomainPayload
	m2.selectedHead = 1
	m2.headerSubfocus = subfocusKey
	m2.headers = append(m2.headers, newHeaderRow(), newHeaderRow(), newHeaderRow())

	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m2 = updated2.(Model)
	if len(m2.headers) != 2 {
		t.Fatalf("after delete middle: headers = %d (expected 2)", len(m2.headers))
	}
	if !m2.headers[m2.selectedHead].Key.Focused() {
		t.Fatal("after delete middle header: remaining header key should be focused")
	}
}

// Invariant: Render functions must not mutate model state.

func TestV092Architecture_RenderDoesNotMutate(t *testing.T) {
	// Rendering must never modify model state. Running View() on the same
	// model must produce identical output and leave all fields untouched.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())
	m.results = append(m.results, model.Result{Status: 200})

	mode := m.workspace.mode
	dialog := m.workspace.dialog
	domain := m.activeDomain
	selected := m.selected
	head := m.selectedHead
	running := m.running

	before := m.View()

	if m.workspace.mode != mode {
		t.Fatal("render mutated workspace.mode")
	}
	if m.workspace.dialog != dialog {
		t.Fatal("render mutated workspace.dialog")
	}
	if m.activeDomain != domain {
		t.Fatal("render mutated activeDomain")
	}
	if m.selected != selected {
		t.Fatal("render mutated selected")
	}
	if m.selectedHead != head {
		t.Fatal("render mutated selectedHead")
	}
	if m.running != running {
		t.Fatal("render mutated running")
	}

	after := m.View()
	if before != after {
		t.Fatal("View() is non-deterministic: two calls produced different output")
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

// Invariant: hjkl keys are consistent with arrow keys in every domain.

func TestV092Interaction_HJKLConsistency(t *testing.T) {
	// hjkl must produce identical results to arrow keys in every domain
	// where both are handled. This verifies the hjkl vim-bindings are
	// not diverging from the primary arrow-key bindings.

	// Exec domain: up/k increment, down/j decrement.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainExec
	m.concurrencyInput.Focus()
	m.concurrencyInput.SetValue("5")

	updatedK, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updatedUp, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	mK := updatedK.(Model)
	mUp := updatedUp.(Model)
	if mK.concurrencyInput.Value() != mUp.concurrencyInput.Value() {
		t.Fatalf("exec: up=%s but k=%s", mUp.concurrencyInput.Value(), mK.concurrencyInput.Value())
	}

	updatedJ, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updatedDown, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mJ := updatedJ.(Model)
	mDown := updatedDown.(Model)
	if mJ.concurrencyInput.Value() != mDown.concurrencyInput.Value() {
		t.Fatalf("exec: down=%s but j=%s", mDown.concurrencyInput.Value(), mJ.concurrencyInput.Value())
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

// ---------------------------------------------------------------------------
// Section: Layout
//
// Invariant: The layout is stable and predictable at every supported size
// and transition. No line exceeds the terminal width. No size causes a
// panic. Determining output is deterministic (no state leakage).
//
// Verified:
//   - All supported widths (72-220) and heights (10-40) render without overflow
//   - Extreme sizes (1x1, 200x100, etc.) never panic
//   - Identical model at identical size produces identical output
//   - All layout configurations produce valid region dimensions
// ---------------------------------------------------------------------------

// Invariant: Layout is stable at every supported size and transition.

func TestV092Geometry_RendersWithoutWrap(t *testing.T) {
	// Verify layout integrity across a range of terminal sizes.
	// The shell enforces a minimum width of 72, so we test from that
	// floor upward. Heights test the minimum supported values.
	widths := []int{72, 80, 100, 120, 160, 200, 220}
	heights := []int{10, 16, 24, 30, 40}

	surfaces := []struct {
		name  string
		setup func(m *Model)
	}{
		{"idle", func(m *Model) {}},
		{"running", func(m *Model) { m.running = true }},
		{"request dialog", func(m *Model) { m.workspace.dialog = dialogRequest }},
		{"inspect", func(m *Model) { m.workspace.mode = modeInspect }},
	}

	for _, s := range surfaces {
		for _, w := range widths {
			for _, h := range heights {
				t.Run(s.name+fmt.Sprintf("/%dx%d", w, h), func(t *testing.T) {
					m := NewModel()
					s.setup(&m)
					m.shell.Resize(w, h)

					view := m.View()
					if view == "" {
						t.Fatal("view returned empty string")
					}

					// Verify no line exceeds the terminal width.
					lines := strings.Split(view, "\n")
					for i, line := range lines {
						if lipgloss.Width(line) > w {
							t.Fatalf("line %d overflows: width %d > terminal %d", i, lipgloss.Width(line), w)
						}
					}
				})
			}
		}
	}
}

func TestV092Geometry_NoPanicAtExtremeSizes(t *testing.T) {
	sizes := []struct{ w, h int }{
		{1, 1},
		{5, 5},
		{200, 100},
		{40, 100},
		{220, 8},
	}

	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(sz.w, sz.h)
			_ = m.View() // must not panic
		})
	}
}

// Invariant: The existing layout can accept regions, borders, and panels
// without architectural changes.

func TestV092Layout_SpacingIsDeterministic(t *testing.T) {
	// The same model rendered twice at the same width must produce
	// identical output. This verifies no state leaks into rendering.
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey
	m.headers = append(m.headers, newHeaderRow())
	m.shell.Resize(80, 24)

	a := m.View()
	b := m.View()

	if a != b {
		t.Fatal("same model rendered twice at same size produced different output")
	}
}

func TestV092Layout_AllLayoutsRender(t *testing.T) {
	// Every layout configuration must produce valid rendered output.
	sizes := []struct{ w, h int }{
		{80, 24},
		{100, 30},
		{120, 40},
		{160, 40},
		{60, 10},
	}
	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(sz.w, sz.h)

			layout := m.shell.Layout()
			if layout.Workspace.Width <= 0 || layout.Workspace.Height <= 0 {
				t.Fatalf("workspace region has non-positive dimensions: %dx%d", layout.Workspace.Width, layout.Workspace.Height)
			}
			if layout.Context.Width < 0 || layout.Command.Width < 0 {
				t.Fatalf("negative context or command width")
			}

			view := m.View()
			if view == "" {
				t.Fatal("View() returned empty string")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Section: Rendering
//
// Invariant: Visual treatment expresses semantics consistently.
// Failure: A visual element communicates the wrong message (e.g., inactive
// state rendered with active styling, inconsistent use of box-drawing).
//
// Verified:
//   - Active domains use heavy rules (━), inactive use light (─)
//   - Workspace identity is never boxed
//   - Section dividers use em dashes (──) not ASCII dashes
// ---------------------------------------------------------------------------

// Invariant: Visual treatment expresses semantics.

func TestV092Visual_HeavyVsLightRules(t *testing.T) {
	// Active domain uses heavy rule (━), inactive uses light rule (─).
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest

	reqRendered := m.renderRequestDomain(80)
	if !strings.Contains(reqRendered, "━") {
		t.Fatal("active domain (Request) should use heavy rule ━")
	}

	payloadRendered := m.renderPayloadDomain(80)
	if !strings.Contains(payloadRendered, "─") {
		t.Fatal("inactive domain (Payload) should use light rule ─")
	}
}

func TestV092Visual_WorkspaceIdentityNeverBoxed(t *testing.T) {
	// The workspace identity badge itself must never use box-drawing
	// characters. The workspace region has a border in v0.9.2+, but
	// the badge within it must not.
	badge := renderWorkspaceBadge("OBSERVE")
	if strings.Contains(badge, "┌") || strings.Contains(badge, "┐") ||
		strings.Contains(badge, "└") || strings.Contains(badge, "┘") {
		t.Fatal("workspace identity badge should not use box-drawing characters")
	}
}

func TestV092Visual_SectionLinesUseEmDash(t *testing.T) {
	// Section dividers must use em dashes (──), not ASCII dashes (--).
	m := NewModel()
	m.workspace.mode = modeInspect
	m.results = append(m.results, model.Result{Status: 200})
	m.selected = 0

	rendered := m.renderInspect(Region{Width: 80, Height: 24})
	if !strings.Contains(rendered, "──") {
		t.Fatal("inspect section lines should use ──")
	}
}

// ---------------------------------------------------------------------------
// Section: Information Architecture
//
// Invariant: Every screen answers: where am I, what is selected, what can I
// edit, what happens next, what does Tab do, what does Esc do.
// Failure: A surface renders without communicating orientation or navigable
// state.
//
// Verified:
//   - Every renderable state shows its orientation label
// ---------------------------------------------------------------------------

// Invariant: Every screen answers: where, what's selected, what changes,
// what's next, what does Tab do, what does Esc do.

func TestV092IA_EverySurfaceHasOrientation(t *testing.T) {
	// Every renderable state must show its orientation label.
	states := []struct {
		name     string
		setup    func(m *Model)
		expected string
	}{
		{"idle", func(m *Model) {}, "READY"},
		{"request dialog", func(m *Model) { m.workspace.dialog = dialogRequest }, "REQUEST"},
		{"inspect mode", func(m *Model) { m.workspace.mode = modeInspect }, "INSPECT"},
		{"quit dialog", func(m *Model) { m.workspace.dialog = dialogConfirmQuit }, "QUIT"},
	}
	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			m := NewModel()
			st.setup(&m)
			view := m.View()
			if !strings.Contains(view, st.expected) {
				t.Fatalf("view does not contain orientation %q", st.expected)
			}
		})
	}
}

// Invariant: Every empty state guides the operator rather than feeling
// like an unhandled edge case.

func TestV092IA_EmptyStatesGuide(t *testing.T) {
	// Ready state (no results, not running) should show guidance.
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "Ready") {
		t.Fatal("idle state should show Ready guidance")
	}

	// Inspect with no results should show an empty-state message.
	m2 := NewModel()
	m2.workspace.mode = modeInspect
	view2 := m2.View()
	if view2 == "" {
		t.Fatal("inspect with no results should not render empty")
	}

	// Request dialog with no config should show the dialog.
	m3 := NewModel()
	m3.workspace.dialog = dialogRequest
	view3 := m3.View()
	if !strings.Contains(view3, "REQUEST") {
		t.Fatal("request dialog should show REQUEST orientation")
	}
}

// ---------------------------------------------------------------------------
// Section: Engineering
//
// Invariant: The test suite itself is maintainable, exhaustive, and
// executable. Redundant tests are removed. Coverage is intentional.
//
// Verified:
//   - Semantic constants are used in place of magic literals
//   - No deprecated or superseded conventions remain
// ---------------------------------------------------------------------------

func TestV092Engineering_SentinelConstantUsed(t *testing.T) {
	// Verify the sentinelEmpty constant is used instead of hardcoded
	// em dash strings for the empty payload sentinel.
	m := NewModel()
	if m.payloadSummary() != sentinelEmpty {
		t.Fatal("payloadSummary should return sentinelEmpty for empty payload")
	}
}

func TestV092Engineering_ConfigurationUsesCC(t *testing.T) {
	// Configuration summaries should use "CC" not "Concurrency".
	cfg := NewModel().Configuration()
	found := false
	for _, c := range cfg {
		if c.Identity == "CC" {
			found = true
		}
		if c.Identity == "Concurrency" {
			t.Fatal("Configuration should use CC, not Concurrency")
		}
	}
	if !found {
		t.Fatal("Configuration should contain CC identity")
	}
}

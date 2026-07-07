package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/divijg19/Pulse/internal/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func keyMsgRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func keyMsgKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func typeName(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

// ---------------------------------------------------------------------------
// Constitutional Audit
// ---------------------------------------------------------------------------

func TestV091Constitution_OrientationDelegates(t *testing.T) {
	m := NewModel()
	if got := orientationLabel(m); got != "READY" {
		t.Fatalf("orientationLabel = %q, want READY", got)
	}
}

func TestV091Constitution_SurfacesAreNamedTypes(t *testing.T) {
	m := NewModel()

	s := m.resolveSurface()
	if _, ok := s.(ReadySurface); !ok {
		t.Fatalf("idle surface should be ReadySurface, got %T", s)
	}

	m2 := NewModel()
	m2.workspace.dialog = dialogRequest
	s2 := m2.resolveSurface()
	if _, ok := s2.(RequestSurface); !ok {
		t.Fatalf("request surface should be RequestSurface, got %T", s2)
	}

	m3 := NewModel()
	m3.workspace.mode = modeInspect
	m3.results = testResults(1)
	s3 := m3.resolveSurface()
	if _, ok := s3.(InspectSurface); !ok {
		t.Fatalf("inspect surface should be InspectSurface, got %T", s3)
	}
}

// ---------------------------------------------------------------------------
// Interaction Ownership Audit
// ---------------------------------------------------------------------------

func TestV091Ownership_ArrowKeysNeverProduceTextInRequestDomain(t *testing.T) {
	m := newRequestModel()
	for _, k := range []tea.KeyType{tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight} {
		urlBefore := m.urlInput.Value()
		m2, _ := m.Update(tea.KeyMsg{Type: k})
		m = m2.(Model)
		if m.urlInput.Value() != urlBefore {
			t.Fatalf("arrow key %v should not modify URL value: got %q", k, m.urlInput.Value())
		}
	}
}

func TestV091Ownership_VimKeysTypeInEditableFields(t *testing.T) {
	// In editable fields, h/j/k/l must type into the focused widget, never navigate.
	m := newRequestModel()

	// 'k' on URL field should type "k", not navigate to Method.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m2m := m2.(Model)
	if m2m.requestField != reqFieldURL {
		t.Fatal("'k' on URL should stay in URL field, not navigate")
	}
	if !strings.HasSuffix(m2m.urlInput.Value(), "k") {
		t.Fatal("'k' on URL should insert 'k' into urlInput, got", m2m.urlInput.Value())
	}

	// Arrows still navigate between fields.
	m3, _ := m2m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3m := m3.(Model)
	if m3m.requestField != reqFieldMethod {
		t.Fatal("↑ on URL should navigate to Method field")
	}
}

func TestV091Ownership_ArrowKeysInPayloadDomain(t *testing.T) {
	m := newRequestPayloadModel()
	m.activeDomain = DomainPayload
	m.selectedHead = 0
	m.headerSubfocus = subfocusKey

	// ↑ at row 0 should cross to DomainRequest
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.activeDomain != DomainRequest {
		t.Fatal("↑ at first header row should cross to Request domain")
	}

	// ↓ at URL should cross back to Payload
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated2.(Model)
	if m3.activeDomain != DomainPayload {
		t.Fatal("↓ at URL should cross back to Payload domain")
	}
}

func TestV091Ownership_ArrowKeysTraversePayloadBodyBoundary(t *testing.T) {
	m := newRequestPayloadModel()
	m.activeDomain = DomainPayload
	m.selectedHead = len(m.headers) - 1 // last header row
	m.headerSubfocus = subfocusValue

	// ↓ at last header value should go to body
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(Model)
	if m2.selectedHead != bodyFocus {
		t.Fatal("↓ at last header row should advance to body")
	}
	if m2.activeDomain != DomainPayload {
		t.Fatal("↓ at last header value should stay in Payload domain")
	}

	// ↑ at body should stay in body (editor cursor move, not navigation)
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3 := updated2.(Model)
	if m3.selectedHead != bodyFocus {
		t.Fatal("↑ at body should stay in body (move editor cursor, not navigate)")
	}
	if m3.activeDomain != DomainPayload {
		t.Fatal("↑ at body should stay in Payload domain")
	}
}

func TestV091Ownership_ExecDomainArrowAdjustsConcurrency(t *testing.T) {
	m := newRequestExecModel()
	m.concurrencyInput.SetValue("5")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.concurrencyInput.Value() != "6" {
		t.Fatalf("↑ in exec domain should increment concurrency: got %q", m2.concurrencyInput.Value())
	}
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated2.(Model)
	if m3.concurrencyInput.Value() != "5" {
		t.Fatalf("↓ in exec domain should decrement concurrency: got %q", m3.concurrencyInput.Value())
	}
}

func TestV091Ownership_FocusedGuardPreventsGhostTyping(t *testing.T) {
	m := newRequestModel()
	m.activeDomain = DomainRequest
	m.requestField = reqFieldURL
	m.urlInput.Focus()
	m.concurrencyInput.Blur()
	ccBefore := m.concurrencyInput.Value()

	// 'j' on URL field should type into urlInput, stay in request domain,
	// and not ghost-type into concurrencyInput.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m2 := updated.(Model)
	if !strings.HasSuffix(m2.urlInput.Value(), "j") {
		t.Fatal("'j' on URL should insert 'j' into urlInput, got", m2.urlInput.Value())
	}
	if m2.activeDomain != DomainRequest {
		t.Fatal("'j' on URL should not change domain (should stay in request)")
	}
	if m2.concurrencyInput.Value() != ccBefore {
		t.Fatal("'j' in request domain with URL focused should not modify concurrencyInput")
	}
}

func TestV091Ownership_AdvanceDomainGranularTab(t *testing.T) {
	m := newRequestModel()

	// Tab: URL → Payload (Key, row 0)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m2 := updated.(Model)
	if m2.activeDomain != DomainPayload || m2.selectedHead != 0 || m2.headerSubfocus != subfocusKey {
		t.Fatal("Tab from URL should go to Payload header Key")
	}

	// Tab: Key → Value
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	m3 := updated2.(Model)
	if m3.activeDomain != DomainPayload || m3.headerSubfocus != subfocusValue {
		t.Fatal("Tab from header Key should go to header Value")
	}

	// Tab: Value → Body
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	m4 := updated3.(Model)
	if m4.selectedHead != bodyFocus {
		t.Fatal("Tab from header Value should go to body")
	}

	// Tab: Body → Exec
	updated4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyTab})
	m5 := updated4.(Model)
	if m5.activeDomain != DomainExec {
		t.Fatal("Tab from body should go to Exec domain")
	}
}

// ---------------------------------------------------------------------------
// Architecture: Workspace, Surfaces, Domains, Focus, Render
// ---------------------------------------------------------------------------

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

	label := orientationLabel(m)
	if label != "READY" {
		t.Fatalf("orientationLabel = %q (expected READY)", label)
	}

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

func TestV092Focus_EmptyHeaderGuard(t *testing.T) {
	m := NewModel()
	m.headers = nil
	m.selectedHead = 0

	m.focusPayloadKey()
	m.focusPayloadValue()

	m.selectedHead = 5
	m.focusPayloadKey()
	m.focusPayloadValue()

	m.selectedHead = -10
	m.focusPayloadKey()
	m.focusPayloadValue()
}

func TestV092Focus_TransitionConsistency(t *testing.T) {
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

func TestV092Architecture_RenderDoesNotMutate(t *testing.T) {
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

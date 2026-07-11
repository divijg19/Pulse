package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

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

// ---------------------------------------------------------------------------
// v0.9.9 — Comparison Invariants
// ---------------------------------------------------------------------------

func TestV099Comparing_StateComparingImpliesAnalysis(t *testing.T) {
	setup := func(results int) Model {
		m := NewModel()
		m.shell.Resize(100, 30)
		m.results = testResults(results)
		return m
	}

	// Path 1: observe mode, BaselineMarked → Comparing
	t.Run("observe baseline marked", func(t *testing.T) {
		m := setup(3)
		m.selected = 1
		m.workspace.mode = modeObserve
		m.workspace.compare.Baseline = &m.results[0]
		m.workspace.compare.State = CompareBaselineMarked

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		m2m := m2.(Model)
		if m2m.workspace.compare.State == CompareComparing && m2m.workspace.compare.Analysis == nil {
			t.Fatal("H1: State=Comparing but Analysis=nil after observe baseline marked")
		}
	})

	// Path 2: observe mode, Comparing → Comparing (replace candidate)
	t.Run("observe replace candidate", func(t *testing.T) {
		m := setup(4)
		m.selected = 2
		m.workspace.mode = modeCompare
		m.workspace.compare.Baseline = &m.results[0]
		m.workspace.compare.Candidate = &m.results[1]
		m.workspace.compare.State = CompareComparing
		m.workspace.compare.Reference = &m.results[0]
		m.workspace.compare.refreshAnalysis()

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		m2m := m2.(Model)
		if m2m.workspace.compare.State == CompareComparing && m2m.workspace.compare.Analysis == nil {
			t.Fatal("State=Comparing but Analysis=nil after observe replace candidate")
		}
	})

	// Path 3: inspect mode, Idle → Comparing with Reference
	t.Run("inspect idle with pin", func(t *testing.T) {
		m := setup(3)
		m.selected = 1
		m.workspace.mode = modeInspect
		m.workspace.compare.Reference = &m.results[0]
		m.workspace.compare.State = CompareIdle

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		m2m := m2.(Model)
		if m2m.workspace.compare.State == CompareComparing && m2m.workspace.compare.Analysis == nil {
			t.Fatal("State=Comparing but Analysis=nil after inspect idle with pin")
		}
	})

	// Path 4: inspect mode, BaselineMarked → Comparing
	t.Run("inspect baseline marked", func(t *testing.T) {
		m := setup(3)
		m.selected = 1
		m.workspace.mode = modeInspect
		m.workspace.compare.Baseline = &m.results[0]
		m.workspace.compare.State = CompareBaselineMarked

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		m2m := m2.(Model)
		if m2m.workspace.compare.State == CompareComparing && m2m.workspace.compare.Analysis == nil {
			t.Fatal("State=Comparing but Analysis=nil after inspect baseline marked")
		}
	})

	// Path 5: swap in compare mode
	t.Run("swap in compare", func(t *testing.T) {
		m := setup(3)
		m.workspace.mode = modeCompare
		m.workspace.compare.Reference = &m.results[0]
		m.workspace.compare.Baseline = &m.results[0]
		m.workspace.compare.Candidate = &m.results[1]
		m.workspace.compare.State = CompareComparing
		m.workspace.compare.refreshAnalysis()

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		m2m := m2.(Model)
		if m2m.workspace.compare.State != CompareComparing {
			t.Fatal("swap should remain in Comparing state")
		}
		if m2m.workspace.compare.Analysis == nil {
			t.Fatal("swap should recompute Analysis, got nil")
		}
	})
}

func TestV099Comparing_PinInvariant(t *testing.T) {
	// Reference == nil on a fresh Model
	m := NewModel()
	if m.workspace.compare.Reference != nil {
		t.Fatal("fresh Model should have nil Reference")
	}
	if m.workspace.compare.State == CompareComparing {
		t.Fatal("fresh Model should not be in Comparing state")
	}
}

func TestV099Architecture_EngineNoRenderImports(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{.Imports}}", "comparison.go")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"lipgloss", "bubbletea", "ansi", "termenv"}
	for _, pkg := range forbidden {
		if strings.Contains(string(out), pkg) {
			t.Errorf("comparison.go imports rendering package: %s", pkg)
		}
	}
}

func TestV099Status_ClassifyStatusCoverage(t *testing.T) {
	tt := []struct {
		status int
		class  StatusClass
	}{
		{-1, StatusUnknown},
		{0, StatusUnknown},
		{50, StatusUnknown},
		{99, StatusUnknown},
		{100, StatusInfo},
		{101, StatusInfo},
		{199, StatusInfo},
		{200, StatusSuccess},
		{201, StatusSuccess},
		{299, StatusSuccess},
		{300, StatusRedirect},
		{301, StatusRedirect},
		{399, StatusRedirect},
		{400, StatusClientError},
		{404, StatusClientError},
		{499, StatusClientError},
		{500, StatusServerError},
		{502, StatusServerError},
		{599, StatusServerError},
		{600, StatusServerError},
	}
	for _, tc := range tt {
		got := ClassifyStatus(tc.status)
		if got != tc.class {
			t.Errorf("ClassifyStatus(%d) = %d (expected %d)", tc.status, got, tc.class)
		}
	}
}

func TestV099Status_ResultStatusRoundtrip(t *testing.T) {
	tt := []struct {
		status int
		want   string
	}{
		{0, "ERR"},
		{101, "101 Info"},
		{199, "199 Info"},
		{200, "200 OK"},
		{201, "201 OK"},
		{301, "301 Redirect"},
		{302, "302 Redirect"},
		{400, "400"},
		{404, "404"},
		{500, "500"},
		{50, "50"},
		{99, "99"},
	}
	for _, tc := range tt {
		got := resultStatus(model.Result{Status: tc.status})
		if got != tc.want {
			t.Errorf("resultStatus(%d) = %q (expected %q)", tc.status, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// v0.10.0 — Compare Rendering & Behavioural Invariants
// ---------------------------------------------------------------------------

func verifySectionOrdering(t *testing.T, out string) {
	whyIdx := strings.Index(out, "── WHY ──")
	evIdx := strings.Index(out, "── EVIDENCE ──")
	dtIdx := strings.Index(out, "── DETAILS ──")

	if whyIdx != -1 && evIdx != -1 {
		if whyIdx > evIdx {
			t.Errorf("Semantic order violated: EVIDENCE rendered before WHY:\n%s", out)
		}
	}
	if evIdx != -1 && dtIdx != -1 {
		if evIdx > dtIdx {
			t.Errorf("Semantic order violated: DETAILS rendered before EVIDENCE:\n%s", out)
		}
	}
	if whyIdx != -1 && dtIdx != -1 {
		if whyIdx > dtIdx {
			t.Errorf("Semantic order violated: DETAILS rendered before WHY:\n%s", out)
		}
	}
}

func TestV0100_CompareBehaviourScenarios(t *testing.T) {
	scenarios := []struct {
		name         string
		baseline     model.Result
		candidate    model.Result
		wantVerdict  string
		wantWhys     []string
		wantEvidence []string
		wantDetails  []string
	}{
		{
			name:         "identical",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseBody: "ok"},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseBody: "ok"},
			wantVerdict:  "No significant changes",
			wantWhys:     nil,
			wantEvidence: nil,
			wantDetails:  nil,
		},
		{
			name:         "improved",
			baseline:     model.Result{Status: 500, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			wantVerdict:  "Improvement detected",
			wantWhys:     []string{"Status improved", "Latency decreased"},
			wantEvidence: []string{"Status: 500 → 200", "Latency: 100ms → 50ms"},
			wantDetails:  nil,
		},
		{
			name:         "regressed",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 500, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/a"},
			wantVerdict:  "Regression detected",
			wantWhys:     []string{"Status regressed", "Latency increased"},
			wantEvidence: []string{"Status: 200 → 500", "Latency: 50ms → 100ms"},
			wantDetails:  nil,
		},
		{
			name:         "mixed",
			baseline:     model.Result{Status: 500, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 200, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/a"},
			wantVerdict:  "Mixed results",
			wantWhys:     []string{"Status improved", "Latency increased"},
			wantEvidence: []string{"Status: 500 → 200", "Latency: 50ms → 100ms"},
			wantDetails:  nil,
		},
		{
			name:         "metadata-only",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 200, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/b"},
			wantVerdict:  "Regression detected",
			wantWhys:     []string{"Latency increased"},
			wantEvidence: []string{"Latency: 50ms → 100ms"},
			wantDetails:  []string{"URL:", "https://example.com/a", "https://example.com/b"},
		},
		{
			name:         "header-only",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseHeaders: map[string]string{"X-Test": "foo"}},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseHeaders: map[string]string{"X-Test": "bar"}},
			wantVerdict:  "No significant changes",
			wantWhys:     []string{"Headers changed"},
			wantEvidence: []string{"Headers: 1 header(s) differ"},
		},
		{
			name:         "body-only",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseBody: "hello"},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", ResponseBody: "world"},
			wantVerdict:  "No significant changes",
			wantWhys:     []string{"Body changed"},
			wantEvidence: []string{"Body: 5 bytes (no change), 2 line(s) differ"},
		},
		{
			name:         "status-only",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 400, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			wantVerdict:  "Regression detected",
			wantWhys:     []string{"Status regressed"},
			wantEvidence: []string{"Status: 200 → 400"},
		},
		{
			name:         "latency-only",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"},
			candidate:    model.Result{Status: 200, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/a"},
			wantVerdict:  "Regression detected",
			wantWhys:     []string{"Latency increased"},
			wantEvidence: []string{"Latency: 50ms → 100ms"},
		},
		{
			name:         "error transitions",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", Error: "old error"},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", Error: ""},
			wantVerdict:  "Improvement detected",
			wantWhys:     []string{"Error resolved"},
			wantEvidence: []string{"Error: old error → (resolved)"},
		},
		{
			name:         "timeout transitions",
			baseline:     model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", Error: ""},
			candidate:    model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a", Error: "timeout"},
			wantVerdict:  "Regression detected",
			wantWhys:     []string{"Error introduced"},
			wantEvidence: []string{"Error: (none) → timeout"},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			m := NewModel()
			m.results = []model.Result{sc.baseline, sc.candidate}
			setComparing(&m, &sc.baseline, &sc.candidate)

			m.workspace.compare.View = CompareViewOverview
			outOverview := m.renderCompare(Region{Width: 100, Height: 30})

			m.workspace.compare.View = CompareViewEvidence
			outEvidence := m.renderCompare(Region{Width: 100, Height: 30})

			m.workspace.compare.View = CompareViewDiff
			outDetails := m.renderCompare(Region{Width: 100, Height: 30})

			// The semantic order (WHY → EVIDENCE → DETAILS) must hold across the
			// three views taken together; renderCompare renders one view per call
			// so the ordering can only be asserted on the concatenation.
			verifySectionOrdering(t, outOverview+"\n"+outEvidence+"\n"+outDetails)

			if !strings.Contains(outOverview, sc.wantVerdict) {
				t.Errorf("expected verdict %q, got output:\n%s", sc.wantVerdict, outOverview)
			}

			for _, why := range sc.wantWhys {
				if !strings.Contains(outOverview, why) {
					t.Errorf("expected why message %q to be rendered, got output:\n%s", why, outOverview)
				}
			}

			for _, ev := range sc.wantEvidence {
				if !strings.Contains(outEvidence, ev) {
					t.Errorf("expected evidence %q to be rendered, got output:\n%s", ev, outEvidence)
				}
			}

			for _, dt := range sc.wantDetails {
				if !strings.Contains(outDetails, dt) {
					t.Errorf("expected detail %q to be rendered, got output:\n%s", dt, outDetails)
				}
			}
		})
	}
}

func TestV0100_CompareResponsiveInvariants(t *testing.T) {
	m := NewModel()
	baseline := model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"}
	candidate := model.Result{Status: 500, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/b"}
	m.results = []model.Result{baseline, candidate}
	setComparing(&m, &baseline, &candidate)

	m.workspace.compare.View = CompareViewEvidence
	widths := []int{80, 100, 119, 120, 150}
	for _, w := range widths {
		t.Run(fmt.Sprintf("width_%d", w), func(t *testing.T) {
			out := m.renderCompare(Region{Width: w, Height: 30})

			if strings.Contains(out, "│") {
				t.Errorf("Wide layout must not contain '│' column separator anymore:\n%s", out)
			}
		})
	}

	t.Run("narrow_79", func(t *testing.T) {
		out := m.renderCompare(Region{Width: 79, Height: 30})
		if !strings.Contains(out, "requires at least 80 columns") {
			t.Errorf("Narrow layout under 80 must require at least 80 columns, got:\n%s", out)
		}
	})
}

func TestV0100_CompareRendererConsumesAnalysisOnly(t *testing.T) {
	m := NewModel()
	baseline := model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"}
	candidate := model.Result{Status: 500, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/b"}
	m.results = []model.Result{baseline, candidate}
	setComparing(&m, &baseline, &candidate)

	baseOutput := m.renderCompare(Region{Width: 100, Height: 30})

	m.activeDomain = DomainPayload
	m.running = true
	m.requestField = reqFieldMethod
	m.selected = 4
	m.inspectZone = zoneWhy

	mutatedOutput := m.renderCompare(Region{Width: 100, Height: 30})

	if baseOutput != mutatedOutput {
		t.Error("Renderer does not consume structured analysis only; mutating unrelated model fields changed the comparison view")
	}
}

func TestV0100_CompareRenderDoesNotMutate(t *testing.T) {
	m := NewModel()
	baseline := model.Result{Status: 200, Latency: 50 * time.Millisecond, RequestURL: "https://example.com/a"}
	candidate := model.Result{Status: 500, Latency: 100 * time.Millisecond, RequestURL: "https://example.com/b"}
	m.results = []model.Result{baseline, candidate}
	setComparing(&m, &baseline, &candidate)

	activeDomain := m.activeDomain
	running := m.running
	requestField := m.requestField
	selected := m.selected
	inspectZone := m.inspectZone
	compareState := m.workspace.compare.State

	_ = m.renderCompare(Region{Width: 100, Height: 30})

	if m.activeDomain != activeDomain ||
		m.running != running ||
		m.requestField != requestField ||
		m.selected != selected ||
		m.inspectZone != inspectZone ||
		m.workspace.compare.State != compareState {
		t.Error("renderCompare mutated the model state during rendering")
	}
}

// ---------------------------------------------------------------------------
// v0.10.2 — Compare Workflow Convergence Tests
// ---------------------------------------------------------------------------

// enterComparing marks results[baselineIdx] as baseline and compares it against
// results[candidateIdx], returning the Model in the Comparing state. It factors
// out the repeated two-press sequence shared by the workflow tests.
func enterComparing(t *testing.T, baselineIdx, candidateIdx int) Model {
	t.Helper()
	m := NewModel()
	m.results = testResults(3)
	m.selected = baselineIdx
	updated, _ := m.Update(keyMsgRune('c'))
	m = updated.(Model)
	m.selected = candidateIdx
	updated, _ = m.Update(keyMsgRune('c'))
	return updated.(Model)
}

// setComparing wires a Model into the Comparing state with the given baseline
// and candidate, recomputing the analysis exactly once — mirroring how the
// workspace reaches that state through MarkBaseline + SelectCandidate.
func setComparing(m *Model, baseline, candidate *model.Result) {
	m.workspace.compare.Baseline = baseline
	m.workspace.compare.Candidate = candidate
	m.workspace.compare.State = CompareComparing
	m.workspace.compare.Reference = baseline
	m.workspace.compare.refreshAnalysis()
}

func TestV0102Workflow_MarkReplace(t *testing.T) {
	// Mark A as baseline → c on B → Compare (B replaces A as baseline)
	m := NewModel()
	m.results = testResults(3)
	m.selected = 0

	// First c marks baseline
	updated, _ := m.Update(keyMsgRune('c'))
	m = updated.(Model)
	if m.workspace.compare.State != CompareBaselineMarked {
		t.Fatal("first c should mark baseline")
	}
	if !resultsEqual(*m.workspace.compare.Baseline, m.results[0]) {
		t.Fatal("baseline should be results[0]")
	}

	// Second c on different result enters Compare
	m.selected = 2
	updated, _ = m.Update(keyMsgRune('c'))
	m = updated.(Model)
	if m.workspace.compare.State != CompareComparing {
		t.Fatal("c on different result should enter Compare")
	}
	if !resultsEqual(*m.workspace.compare.Candidate, m.results[2]) {
		t.Fatal("candidate should be results[2]")
	}
}

func TestV0102Workflow_ClearPreservesReference(t *testing.T) {
	// Start a comparison, reference it, then clear — Reference survives
	m := enterComparing(t, 0, 1)

	// Reference the baseline (simulate startRun referencing)
	m.workspace.compare.Reference = &m.results[0]

	// Enter Compare mode and press x
	m.workspace.mode = modeCompare
	updated, _ := m.Update(keyMsgRune('x'))
	m = updated.(Model)

	if m.workspace.compare.State != CompareIdle {
		t.Fatal("x should reset session to Idle")
	}
	if m.workspace.compare.Reference == nil {
		t.Fatal("x must preserve Reference")
	}
	if m.workspace.mode != modeObserve {
		t.Fatal("x should return to Observe")
	}
}

func TestV0102Workflow_SwapTwiceReturns(t *testing.T) {
	// Compare A vs B → s → B is baseline, A is candidate → s → back to original
	m := enterComparing(t, 0, 1)

	beforeBaseline := *m.workspace.compare.Baseline
	beforeCandidate := *m.workspace.compare.Candidate

	// Swap
	m.workspace.mode = modeCompare
	updated, _ := m.Update(keyMsgRune('s'))
	m = updated.(Model)

	if !resultsEqual(*m.workspace.compare.Baseline, beforeCandidate) {
		t.Fatal("swap should exchange baseline with candidate")
	}
	if !resultsEqual(*m.workspace.compare.Candidate, beforeBaseline) {
		t.Fatal("swap should exchange candidate with baseline")
	}

	// Swap back
	updated, _ = m.Update(keyMsgRune('s'))
	m = updated.(Model)

	if !resultsEqual(*m.workspace.compare.Baseline, beforeBaseline) {
		t.Fatal("double swap should restore original baseline")
	}
	if !resultsEqual(*m.workspace.compare.Candidate, beforeCandidate) {
		t.Fatal("double swap should restore original candidate")
	}
}

func TestV0102Workflow_ExitPreservesSession(t *testing.T) {
	// Compare → Esc → Observe → session preserved
	m := enterComparing(t, 0, 1)

	beforeBaseline := *m.workspace.compare.Baseline
	beforeCandidate := *m.workspace.compare.Candidate

	// Esc exits compare, preserves session
	m.workspace.mode = modeCompare
	updated, _ := m.Update(keyMsgKey(tea.KeyEsc))
	m = updated.(Model)

	if m.workspace.mode != modeObserve {
		t.Fatal("Esc should return to Observe")
	}
	if !resultsEqual(*m.workspace.compare.Baseline, beforeBaseline) {
		t.Fatal("Esc must preserve baseline")
	}
	if !resultsEqual(*m.workspace.compare.Candidate, beforeCandidate) {
		t.Fatal("Esc must preserve candidate")
	}
	if m.workspace.compare.State != CompareComparing {
		t.Fatal("Esc must preserve session state")
	}
}

func TestV0102Workflow_ResumeFromObserve(t *testing.T) {
	// Compare A vs B → Esc → Observe → c on candidate (▶) → re-enters Compare
	// without disturbing the comparison.
	m := enterComparing(t, 0, 1)
	if m.workspace.mode != modeCompare {
		t.Fatal("should be in Compare after second c")
	}

	// Exit back to Observe; comparison state is preserved.
	m.workspace.mode = modeObserve
	beforeBaseline := *m.workspace.compare.Baseline
	beforeCandidate := *m.workspace.compare.Candidate

	// c on the candidate (marked ▶) must resume the workspace.
	m.selected = 1
	updated, _ := m.Update(keyMsgRune('c'))
	m = updated.(Model)

	if m.workspace.mode != modeCompare {
		t.Fatal("c on candidate should resume Compare from Observe")
	}
	if !resultsEqual(*m.workspace.compare.Baseline, beforeBaseline) {
		t.Fatal("resume must preserve baseline")
	}
	if !resultsEqual(*m.workspace.compare.Candidate, beforeCandidate) {
		t.Fatal("resume must preserve candidate")
	}
}

func TestV0102Workflow_ClearComparison(t *testing.T) {
	// Compare A vs B → x → session cleared, back to Observe
	m := enterComparing(t, 0, 1)

	// x clears
	m.workspace.mode = modeCompare
	updated, _ := m.Update(keyMsgRune('x'))
	m = updated.(Model)

	if m.workspace.compare.State != CompareIdle {
		t.Fatal("x should set session to Idle")
	}
	if m.workspace.mode != modeObserve {
		t.Fatal("x should return to Observe")
	}
}

func TestV0102Workflow_CrossRunReference(t *testing.T) {
	// Mark baseline → startRun → Reference set → session reset
	// → c on new result → compare against reference baseline
	m := NewModel()
	m.results = testResults(3)
	m.selected = 0
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)

	updated, _ := m.Update(keyMsgRune('c'))
	m = updated.(Model)

	// Simulate startRun
	m.workspace.compare.Reference = &m.results[0]
	m.workspace.compare = NewCompareWorkspace()
	m.workspace.compare.Reference = &m.results[0]
	m.results = nil

	// Add new results and compare
	m.results = append(m.results, model.Result{Status: 200, Latency: 10 * time.Millisecond})
	m.selected = 0

	updated, _ = m.Update(keyMsgRune('c'))
	m = updated.(Model)

	if m.workspace.compare.State != CompareComparing {
		t.Fatal("c with reference baseline should enter Compare")
	}
	if m.workspace.compare.Reference == nil {
		t.Fatal("Reference must survive")
	}
}

func TestV0102Workflow_MarkSameUnmarks(t *testing.T) {
	// Baseline marked → c on same result → clears session
	m := enterComparing(t, 0, 0)

	if m.workspace.compare.State != CompareIdle {
		t.Fatal("c on same marked result should clear session")
	}
	if m.workspace.compare.Baseline != nil {
		t.Fatal("c on same marked result should reset baseline")
	}
}

func TestV0102Workflow_NoBaselineCleared(t *testing.T) {
	// With Reference set and no session, render shows Reference state
	m := NewModel()
	m.workspace.compare = NewCompareWorkspace()
	m.workspace.compare.Reference = &model.Result{Status: 200, Latency: 10 * time.Millisecond}

	out := m.renderCompare(Region{Width: 100, Height: 30})
	if !strings.Contains(out, "● Pinned Baseline") {
		t.Fatal("renderCompare with reference should show ● Pinned Baseline, got:\n", out)
	}
}

func TestCompare_RenderRequestDetails(t *testing.T) {
	// Request number (#N) and timestamp must appear for both participants.
	m := compareTestModel()
	for i := range m.results {
		m.results[i].Timestamp = time.Date(2024, 1, 2, 3, 4, 5+i, 0, time.UTC)
	}
	m.workspace.compare.Baseline = &m.results[0]
	m.workspace.compare.Candidate = &m.results[1]
	m.workspace.compare.State = CompareComparing
	m.workspace.compare.refreshAnalysis()
	region := Region{Width: 130, Height: 40}

	m.workspace.compare.View = CompareViewOverview
	outOverview := m.renderCompare(region)
	if !strings.Contains(outOverview, "#001") || !strings.Contains(outOverview, "03:04:05") {
		t.Fatalf("identity should show request number and time, got:\n%s", outOverview)
	}

	m.workspace.compare.View = CompareViewRaw
	outRaw := m.renderCompare(region)
	if !strings.Contains(outRaw, "Request:") {
		t.Fatal("raw view should expose Request details")
	}
	if !strings.Contains(outRaw, "#001") || !strings.Contains(outRaw, "#002") {
		t.Fatal("raw view should show request number for both participants")
	}
}

func TestV0102Workflow_RenounceReference(t *testing.T) {
	// x in Observe with only a reference request renounces persistence.
	m := NewModel()
	m.results = testResults(3)
	m.selected = 0
	updated, _ := m.Update(keyMsgRune('c'))
	m = updated.(Model)

	// Simulate a run that sets the reference for the next run.
	m.workspace.compare.Reference = &m.results[0]
	m.workspace.compare = NewCompareWorkspace()
	m.workspace.compare.Reference = &m.results[0]

	m.workspace.mode = modeObserve
	updated, _ = m.Update(keyMsgRune('x'))
	m = updated.(Model)

	if m.workspace.compare.Reference != nil {
		t.Fatal("x should renounce the reference")
	}
	if m.workspace.compare.State != CompareIdle {
		t.Fatal("workspace should be idle after renounce")
	}
}

func TestCompare_BodyScroll(t *testing.T) {
	// The Body view must scroll through long baseline/candidate bodies.
	m := compareTestModel()
	big := ""
	for i := 0; i < 50; i++ {
		big += fmt.Sprintf("line %d\n", i)
	}
	m.results[0].ResponseBody = big
	m.results[1].ResponseBody = "short candidate"
	m.workspace.compare.Baseline = &m.results[0]
	m.workspace.compare.Candidate = &m.results[1]
	m.workspace.compare.State = CompareComparing
	m.workspace.compare.refreshAnalysis()
	region := Region{Width: 120, Height: 20}

	m.workspace.compare.View = CompareViewBody
	m.inspectBodyOffset = 0
	top := m.renderCompare(region)
	m.inspectBodyOffset = 10
	scrolled := m.renderCompare(region)

	if top == scrolled {
		t.Fatal("body view should change when scrolled")
	}
	if strings.Contains(scrolled, "line 0") {
		t.Fatal("scrolled body should have moved past the first line")
	}
}

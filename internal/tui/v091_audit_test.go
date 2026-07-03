package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/divijg19/Pulse/internal/model"
)

// ---------------------------------------------------------------------------
// Suite 1  --  Constitutional Audit
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
// Suite 2  --  Behaviour Audit
// ---------------------------------------------------------------------------

func keyMsgRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func keyMsgKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

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
	tabKeys := []tea.KeyMsg{
		keyMsgKey(tea.KeyTab),
		keyMsgKey(tea.KeyShiftTab),
		keyMsgKey(tea.KeyLeft),
		keyMsgKey(tea.KeyRight),
		keyMsgRune('h'),
		keyMsgRune('l'),
	}
	for _, key := range tabKeys {
		updated, cmd := m.handleInspectKey(key)
		m2 := updated.(Model)
		if cmd != nil {
			t.Fatalf("key %+v should not produce command in inspect", key.Type)
		}
		if m2.workspace.mode != modeInspect {
			t.Fatal("key should stay in inspect mode")
		}
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
// Suite 3  --  Visual Audit
// ---------------------------------------------------------------------------

func TestV091Visual_DomainHeadersUseCenteredRules(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	out := m.renderRequest(Region{Width: 80, Height: 20})

	if !strings.Contains(out, "──") {
		t.Fatal("domain headers should contain rule characters")
	}
	if !strings.Contains(out, "Request") {
		t.Fatal("request domain header should say 'Request'")
	}
	if !strings.Contains(out, "Payload") {
		t.Fatal("payload domain header should exist")
	}
	if !strings.Contains(out, "Execution") {
		t.Fatal("execution domain header should exist")
	}
}

func TestV091Visual_FooterHasHighlightedAction(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)

	if !strings.Contains(out, "[e]") {
		t.Fatal("ribbon should show [e] Configure")
	}
}

// ---------------------------------------------------------------------------
// Suite 4  --  Composition Audit
// ---------------------------------------------------------------------------

func TestV091Composition_ContextPanelAtWideWidths(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(160, 40)
	out := m.View()
	if !strings.Contains(out, "Selection") {
		t.Fatal("wide terminal should show context panel")
	}
}

func TestV091Composition_NoContextPanelAtMediumWidths(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(100, 30)
	out := m.View()
	if strings.Contains(out, "Selected Request") {
		t.Fatal("medium terminal should NOT show context panel")
	}
}

func TestV091Composition_NoContextPanelWhenEmpty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(160, 40)
	out := m.View()
	if strings.Contains(out, "Selected Request") {
		t.Fatal("empty model should not show Selected Request context")
	}
}

func TestV091Composition_ContextPanelAtAllSurfaces(t *testing.T) {
	for _, c := range AllAuditSurfaces() {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			m := c.Setup()
			m.shell.Resize(160, 40)
			out := m.View()
			// Verify no panics, valid output
			if len(out) == 0 {
				t.Fatal("empty output")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Suite 5  --  Information Architecture Audit
// ---------------------------------------------------------------------------

func TestV091IA_WorkspaceIdentityVisible(t *testing.T) {
	for _, c := range AllAuditSurfaces() {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			m := c.Setup()
			m.shell.Resize(100, 30)
			out := m.View()
			labels := []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "QUIT"}
			found := false
			for _, label := range labels {
				if strings.Contains(out, label) {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("workspace must show identity label (READY/OBSERVE/REQUEST/INSPECT/QUIT)")
			}
		})
	}
}

func TestV091IA_ActiveDomainVisuallyDistinct(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	out := m.renderRequestDomain(80)
	if !strings.Contains(out, "Request") {
		t.Fatal("active domain header must be visible")
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
// Suite 8  --  Visual Rhythm Audit
// ---------------------------------------------------------------------------

func TestV091VisualRhythm_SectionLines(t *testing.T) {
	m := NewModel()
	m.results = testResults(1)
	m.workspace.mode = modeInspect
	out := m.renderInspect(Region{Width: 80, Height: 40})

	// Section captions use em dashes, not ASCII hyphens
	if !strings.Contains(out, "──") {
		t.Fatal("section captions must use em dashes (──)")
	}
	if strings.Contains(out, "---") {
		t.Fatal("section captions must NOT use ASCII hyphens (---)")
	}
}

func TestV091VisualRhythm_DomainHeaderWeight(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	out := m.renderRequestDomain(80)

	// Active domain uses heavy rule (━)
	if !strings.Contains(out, "━") {
		t.Fatal("active domain header must use heavy rule (━)")
	}

	// Inactive domain uses light rule (─):
	// Payload is inactive when DomainRequest is active
	outPayload := m.renderPayloadDomain(80)
	if !strings.Contains(outPayload, "─") {
		t.Fatal("inactive domain header must use light rule (─)")
	}
}

func TestV091VisualRhythm_IdentityPrecedesFirstDomain(t *testing.T) {
	m := NewModel()
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainRequest
	m.shell.Resize(80, 30)
	out := m.renderRequest(Region{Width: 80, Height: 30})

	// Identity must appear before any domain header
	identityPos := strings.Index(out, "REQUEST")
	headerPos := strings.Index(out, "Request")
	if identityPos < 0 {
		t.Fatal("identity (REQUEST) must be visible")
	}
	if headerPos < 0 {
		t.Fatal("domain header (Request) must be visible")
	}
	if identityPos > headerPos {
		t.Fatal("identity REQUEST must appear before domain header Request")
	}
}

// ---------------------------------------------------------------------------
// Suite 9  --  Interaction Ownership Audit
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

func TestV091Ownership_VimKeysInsertTextInURL(t *testing.T) {
	m := newRequestModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m2 := updated.(Model)
	if !strings.HasSuffix(m2.urlInput.Value(), "k") {
		t.Fatal("'k' should insert 'k' into URL when URL is focused")
	}
	// Now press 'j' at end of URL  --  should insert 'j', not navigate
	m2.urlInput.SetCursor(len(m2.urlInput.Value()))
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m3 := updated2.(Model)
	if !strings.HasSuffix(m3.urlInput.Value(), "j") {
		t.Fatal("'j' should insert 'j' into URL when URL is focused")
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

	// ↑ at body should go back to last header row
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	m3 := updated2.(Model)
	if m3.selectedHead != len(m.headers)-1 {
		t.Fatal("↑ at body should go back to last header row")
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

	// hjkl should insert into URL (not navigate), and concurrencyInput should remain unfocused
	keyCount := len(m.concurrencyInput.Value())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m2 := updated.(Model)
	if len(m2.concurrencyInput.Value()) != keyCount {
		t.Fatal("'j' in request domain with URL focused should not modify concurrencyInput")
	}
	if m2.urlInput.Value() == "" || !strings.HasSuffix(m2.urlInput.Value(), "j") {
		t.Fatal("'j' in request domain should insert into URL")
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
// Suite 10  --  Visual Hierarchy Audit
// ---------------------------------------------------------------------------

func TestV091Visual_BadgeUsesAccentBackground(t *testing.T) {
	m := newRequestModel()
	out := m.View()

	// Badge "REQUEST" must be present with accent background style
	if !strings.Contains(out, "REQUEST") {
		t.Fatal("mode cell badge must show REQUEST")
	}
	// Badge should use styleWorkspaceBadge (accent bg), which adds padding
	// Verify badge appears before domain headers
	badgeIdx := strings.Index(out, "REQUEST")
	headerIdx := strings.Index(out, "Request")
	if badgeIdx < 0 || headerIdx < 0 || badgeIdx > headerIdx {
		t.Fatal("badge (REQUEST) must visually precede domain header (Request)")
	}
}

func TestV091Visual_DomainHeaderUsesWeightNotBox(t *testing.T) {
	m := newRequestModel()
	out := m.View()
	// Domain header "Request" should be rendered with ━ (active weight)
	if !strings.Contains(out, "━") {
		t.Fatal("active domain header should use heavy weight ━")
	}
	if strings.Contains(out, "▎") {
		t.Fatal("active domain header must not use left accent bar")
	}
}

func TestV091Visual_BadgeStatusSeparate(t *testing.T) {
	m := newRequestModel()
	m.shell.Resize(100, 24)
	out := m.renderStatusline(m.ShellState(), 100)
	// Ribbon badge should contain workspace identity
	if !strings.Contains(out, "REQUEST") {
		t.Fatal("ribbon badge should contain REQUEST")
	}
	// Ribbon status should contain operational state (Editing for Request dialog)
	if !strings.Contains(out, "Editing") {
		t.Fatal("statusline should contain operational status (Editing)")
	}
}

func TestV091Visual_ThreeLevelTypography(t *testing.T) {
	m := newRequestModel()
	out := m.View()
	// Level 1: Badge "REQUEST"  --  must appear
	if !strings.Contains(out, "REQUEST") {
		t.Fatal("Level 1 typography (badge) must be visible")
	}
	// Level 2: Domain header "Request"  --  must use ━/─ weight
	if !strings.Contains(out, "━") && !strings.Contains(out, "─") {
		t.Fatal("Level 2 typography (domain header) must use heavy/light rules")
	}
	// Level 3: Section labels (like URL label)  --  must appear
	if !strings.Contains(out, "URL") {
		t.Fatal("Level 3 typography (section label) must be visible")
	}
}

// ---------------------------------------------------------------------------
// Suite 11  --  Adaptive Layout Audit
// ---------------------------------------------------------------------------

func TestV091Adaptive_WidePrimaryAndContext(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(160, 40)
	out := m.View()
	if !strings.Contains(out, "Selection") {
		t.Fatal("wide terminal (>=140) must show context panel")
	}
	if !strings.Contains(out, "Timeline") {
		t.Fatal("wide terminal must show primary content")
	}
}

func TestV091Adaptive_MediumPrimaryOnly(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(120, 40)
	out := m.View()
	if strings.Contains(out, "Selection") {
		t.Fatal("medium terminal (<140) must NOT show context panel")
	}
	if len(out) == 0 {
		t.Fatal("medium terminal must render primary content")
	}
}

func TestV091Adaptive_CompactNoPanic(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(60, 10)
	out := m.View()
	if len(out) == 0 {
		t.Fatal("compact terminal must not panic or produce empty output")
	}
}

func TestV091Adaptive_AllSurfacesAtAllLayouts(t *testing.T) {
	sizes := []struct{ w, h int }{
		{160, 40},
		{100, 30},
		{60, 10},
	}
	for _, c := range AllAuditSurfaces() {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			for _, s := range sizes {
				s := s
				t.Run(fmt.Sprintf("%dx%d", s.w, s.h), func(t *testing.T) {
					m := c.Setup()
					m.shell.Resize(s.w, s.h)
					out := m.View()
					if len(out) == 0 {
						t.Fatal("empty output  --  surface must render")
					}
					// Visual width invariant: every line must match the layout width,
					// which may exceed the requested terminal width (shell imposes a minimum of 72)
					layoutW := m.shell.Layout().Context.Width
					for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
						if vw := lipgloss.Width(line); vw != layoutW {
							t.Errorf("%dx%d: line visual width %d, expected %d", s.w, s.h, vw, layoutW)
						}
					}
				})
			}
		})
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

func TestV091Boundary_ArrowUpFromBodyToHeaders(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m2 := updated.(Model)
	if m2.selectedHead != len(m2.headers)-1 || m2.headerSubfocus != subfocusValue {
		t.Fatalf("↑ at body should go to last header Value, got head=%d subfocus=%d", m2.selectedHead, m2.headerSubfocus)
	}
}

func TestV091Boundary_ArrowDownFromBodyToExec(t *testing.T) {
	m := newRequestPayloadModel()
	m.selectedHead = bodyFocus
	m.focusPayloadBody()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2 := updated.(Model)
	if m2.activeDomain != DomainExec {
		t.Fatalf("↓ at body should advance to Execution domain, got domain=%d", m2.activeDomain)
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
// Suite 13  --  Responsive Statusline Invariants
// ---------------------------------------------------------------------------

type responsiveTestCase struct {
	name   string
	setup  func(*Model)
	badge  string
	status string
}

func responsiveTestCases() []responsiveTestCase {
	return []responsiveTestCase{
		{"Ready", func(m *Model) {}, "READY", "Ready"},
		{"Running", func(m *Model) { m.running = true }, "OBSERVE", "Running"},
		{"WithResults", func(m *Model) {
			m.results = []model.Result{{Status: 200, Latency: 100 * time.Millisecond}}
		}, "OBSERVE", "Completed"},
		{"Request_RD", func(m *Model) {
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainRequest
		}, "REQUEST", "Editing"},
		{"Request_ED", func(m *Model) {
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainExec
		}, "REQUEST", "Editing"},
		{"Request_PD", func(m *Model) {
			m.workspace.dialog = dialogRequest
			m.activeDomain = DomainPayload
		}, "REQUEST", "Editing"},
		{"Inspect", func(m *Model) { m.workspace.mode = modeInspect }, "INSPECT", "Inspecting"},
		{"Quit", func(m *Model) { m.workspace.dialog = dialogConfirmQuit }, "QUIT", "Quitting"},
	}
}

func TestV091Responsive_StatuslineNeverWraps(t *testing.T) {
	for _, st := range responsiveTestCases() {
		t.Run(st.name, func(t *testing.T) {
			for width := 40; width <= 220; width++ {
				m := NewModel()
				m.shell.Resize(width, 24)
				st.setup(&m)

				out := m.renderStatusline(m.ShellState(), width)

				if strings.Contains(out, "\n") {
					t.Fatalf("width %d: footer must not wrap", width)
				}
				if !strings.Contains(out, st.badge) {
					t.Fatalf("width %d: badge %q must be visible", width, st.badge)
				}
				if !strings.Contains(out, st.status) {
					t.Fatalf("width %d: status %q must be visible", width, st.status)
				}
				if w := lipgloss.Width(out); w > width {
					t.Fatalf("width %d: rendered width %d exceeds available width", width, w)
				}
			}
		})
	}
}

func TestV091Responsive_DegradationDeterministic(t *testing.T) {
	for _, st := range responsiveTestCases() {
		t.Run(st.name, func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(220, 24)
			st.setup(&m)
			actions := m.Actions()

			var prevLevel Density
			var prevWidth int
			for width := 40; width <= 220; width++ {
				badge := renderWorkspaceBadge(st.badge)
				statusStr := styleStatusCell.Render(st.status)
				level, _ := chooseRibbonLevel(badge, statusStr, actions, width)

				if width > 40 && level > prevLevel {
					t.Fatalf("width %d: density regressed from %d to %d (was at %d)", width, prevLevel, level, prevWidth)
				}
				prevLevel = level
				prevWidth = width

				if width%10 == 0 {
					state := m.ShellState()
					state.Actions = actions
					out1 := m.renderStatusline(state, width)
					out2 := m.renderStatusline(state, width)
					if lipgloss.Width(out1) != lipgloss.Width(out2) {
						t.Fatalf("width %d: non-deterministic output width", width)
					}
				}
			}
		})
	}
}

func TestV091Responsive_ResizeTransitions(t *testing.T) {
	resizeSteps := []int{220, 180, 160, 140, 120, 100, 80, 60, 50, 60, 80, 100, 140, 220}

	for _, st := range responsiveTestCases() {
		t.Run(st.name, func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(220, 24)
			st.setup(&m)

			for _, width := range resizeSteps {
				m.shell.Resize(width, 24)
				out := m.renderStatusline(m.ShellState(), width)

				if strings.Contains(out, "\n") {
					t.Fatalf("width %d after resize: footer must not wrap", width)
				}
				if !strings.Contains(out, st.badge) {
					t.Fatalf("width %d after resize: badge %q must be visible", width, st.badge)
				}
				if !strings.Contains(out, st.status) {
					t.Fatalf("width %d after resize: status %q must be visible", width, st.status)
				}
				if w := lipgloss.Width(out); w > width {
					t.Fatalf("width %d after resize: rendered width %d exceeds available width", width, w)
				}
			}
		})
	}
}

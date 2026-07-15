package tui

import (
	"strings"
	"testing"

	"github.com/divijg19/Pulse/internal/model"
)

// ---------------------------------------------------------------------------
// Visual / Render Tests
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

func TestV091Visual_BadgeUsesAccentBackground(t *testing.T) {
	m := newRequestModel()
	out := m.View().Content

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
	out := m.View().Content
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
	out := m.View().Content
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
// Visual / Render Tests from v092
// ---------------------------------------------------------------------------

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

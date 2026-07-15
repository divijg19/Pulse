package tui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/divijg19/Pulse/internal/model"
)

// ---------------------------------------------------------------------------
// Composition Audit
// ---------------------------------------------------------------------------

func TestLayout_ContextPanelAtWideWidths(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(160, 40)
	out := m.View().Content
	if !strings.Contains(out, "Selection") {
		t.Fatal("wide terminal should show context panel")
	}
}

func TestLayout_NoContextPanelAtMediumWidths(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(100, 30)
	out := m.View().Content
	if strings.Contains(out, "Selected Request") {
		t.Fatal("medium terminal should NOT show context panel")
	}
}

func TestLayout_NoContextPanelWhenEmpty(t *testing.T) {
	m := NewModel()
	m.shell.Resize(160, 40)
	out := m.View().Content
	if strings.Contains(out, "Selected Request") {
		t.Fatal("empty model should not show Selected Request context")
	}
}

// ---------------------------------------------------------------------------
// Adaptive Layout Audit
// ---------------------------------------------------------------------------

func TestLayout_WidePrimaryAndContext(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(160, 40)
	out := m.View().Content
	if !strings.Contains(out, "Selection") {
		t.Fatal("wide terminal (>=140) must show context panel")
	}
	if !strings.Contains(out, "Timeline") {
		t.Fatal("wide terminal must show primary content")
	}
}

func TestLayout_MediumPrimaryOnly(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(120, 40)
	out := m.View().Content
	if strings.Contains(out, "Selection") {
		t.Fatal("medium terminal (<140) must NOT show context panel")
	}
	if len(out) == 0 {
		t.Fatal("medium terminal must render primary content")
	}
}

func TestLayout_CompactNoPanic(t *testing.T) {
	m := newTimelineRunningModel()
	m.shell.Resize(60, 10)
	out := m.View().Content
	if len(out) == 0 {
		t.Fatal("compact terminal must not panic or produce empty output")
	}
}

// ---------------------------------------------------------------------------
// Responsive Statusline Invariants
// ---------------------------------------------------------------------------

type layoutTestCase struct {
	name   string
	setup  func(*Model)
	badge  string
	status string
}

func layoutTestCases() []layoutTestCase {
	return []layoutTestCase{
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

func TestLayout_Responsive_StatuslineNeverWraps(t *testing.T) {
	for _, st := range layoutTestCases() {
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

func TestLayout_Responsive_DegradationDeterministic(t *testing.T) {
	for _, st := range layoutTestCases() {
		t.Run(st.name, func(t *testing.T) {
			m := NewModel()
			m.shell.Resize(220, 24)
			st.setup(&m)
			actions := m.Actions()

			var prevLevel Density
			var prevWidth int
			for width := 40; width <= 220; width++ {
				layout := layoutRibbon(st.badge, st.status, actions, width)

				if width > 40 && layout.Density > prevLevel {
					t.Fatalf("width %d: density regressed from %d to %d (was at %d)", width, prevLevel, layout.Density, prevWidth)
				}
				prevLevel = layout.Density
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

func TestLayout_Responsive_ResizeTransitions(t *testing.T) {
	resizeSteps := []int{220, 180, 160, 140, 120, 100, 80, 60, 50, 60, 80, 100, 140, 220}

	for _, st := range layoutTestCases() {
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

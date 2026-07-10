package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/model"
)

// visualColumnOf returns the 0-indexed visual column of the first occurrence of
// substr in s, counting visible (non-ANSI) runes.
func visualColumnOf(s, substr string) int {
	clean := stripANSI(s)
	idx := strings.Index(clean, substr)
	if idx < 0 {
		return -1
	}
	return idx
}

// ribbonDividerColumn returns the visual column of the single divider glyph "│"
// in the rendered ribbon, or -1 if absent.
func ribbonDividerColumn(out string) int {
	return visualColumnOf(out, "│")
}

// ---------------------------------------------------------------------------
// Footer layout invariants
// ---------------------------------------------------------------------------

func TestFooter_BadgeRegionFixedAndDividerInvariant(t *testing.T) {
	// Every workspace badge occupies exactly ribbonBadgeWidth cells and the
	// divider follows immediately at a fixed column, never moving.
	labels := []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "QUIT", "COMPARE"}

	for _, label := range labels {
		badge := renderWorkspaceBadge(label)
		if w := lipgloss.Width(badge); w != ribbonBadgeWidth {
			t.Fatalf("label %q: badge width %d != fixed ribbonBadgeWidth %d", label, w, ribbonBadgeWidth)
		}
	}

	// Through the full pipeline the divider must sit at the same column for
	// every workspace orientation.
	for _, width := range []int{72, 100, 160} {
		got := -1
		for _, label := range labels {
			m := NewModel()
			switch label {
			case "REQUEST":
				m.workspace.dialog = dialogRequest
			case "INSPECT":
				m.workspace.mode = modeInspect
			case "COMPARE":
				m.workspace.mode = modeCompare
			case "QUIT":
				m.workspace.dialog = dialogConfirmQuit
			}
			out := m.renderStatusline(m.ShellState(), width)
			col := ribbonDividerColumn(out)
			if col != ribbonBadgeWidth {
				t.Fatalf("label %q width %d: divider at column %d, want fixed %d", label, width, col, ribbonBadgeWidth)
			}
			if got == -1 {
				got = col
			} else if col != got {
				t.Fatalf("label %q width %d: divider column %d differs from prior %d", label, width, col, got)
			}
		}
	}
}

func TestFooter_NoDividerBeforeStatus(t *testing.T) {
	// The footer must contain exactly one divider: between badge and actions.
	// There must be no divider before the status region.
	for _, width := range []int{40, 72, 100, 160, 200} {
		m := NewModel()
		m.running = true
		out := m.renderStatusline(m.ShellState(), width)
		if n := strings.Count(stripANSI(out), "│"); n != 1 {
			t.Fatalf("width %d: expected exactly 1 divider, found %d", width, n)
		}
	}
}

func TestFooter_LayoutOrderingInvariant(t *testing.T) {
	// Order is invariant: Badge → Divider → Actions → Status.
	m := NewModel()
	m.running = false
	m.results = []model.Result{{Status: 200}}

	for _, width := range []int{100, 160} {
		out := m.renderStatusline(m.ShellState(), width)

		colBadge := visualColumnOf(out, "OBSERVE")
		colDivider := ribbonDividerColumn(out)
		colActions := visualColumnOf(out, "[")
		colStatus := visualColumnOf(out, "Completed")

		if !(colBadge < colDivider && colDivider < colActions && colActions < colStatus) {
			t.Fatalf("width %d: ordering violated badge=%d divider=%d actions=%d status=%d",
				width, colBadge, colDivider, colActions, colStatus)
		}
	}
}

// ---------------------------------------------------------------------------
// Status truncation regression
// ---------------------------------------------------------------------------

func TestFooter_StatusFinalCharNeverLost(t *testing.T) {
	// The final visible character of every status must be preserved (or
	// replaced by an explicit ellipsis), never silently dropped.
	statuses := []string{
		"Ready",
		"Running",
		"Stopped",
		"Failed",
		"Error",
		"Validation failed: concurrency must be between 1 and 100 and method must be GET or POST",
		"Execution summary: 1240 requests completed in 4.82s with 3 redirects and 2 client errors",
	}

	for _, st := range statuses {
		for _, width := range []int{72, 80, 100, 120, 160, 200} {
			layout := layoutRibbon("OBSERVE", st, []Action{}, width)
			out := renderRibbon(layout)
			clean := stripANSI(out)

			if w := lipgloss.Width(out); w > width {
				t.Fatalf("status %q width %d: rendered %d exceeds terminal width", st, width, w)
			}

			// Either the full status is present, or it was intentionally
			// truncated. When truncated, the ellipsis must be the final visible
			// status character (modulo the fixed right margin, which is trimmed
			// here). The final meaningful character is never silently dropped.
			if strings.Contains(clean, st) {
				continue
			}
			if !strings.HasSuffix(strings.TrimRight(clean, " "), "…") {
				t.Fatalf("status %q width %d: status neither complete nor ending in ellipsis (got %q)", st, width, clean)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Responsive rendering
// ---------------------------------------------------------------------------

func TestFooter_ResponsiveExactWidth(t *testing.T) {
	widths := []int{72, 80, 90, 100, 120, 160, 200}
	orientations := []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "COMPARE", "QUIT"}

	for _, width := range widths {
		for _, orient := range orientations {
			m := NewModel()
			switch orient {
			case "REQUEST":
				m.workspace.dialog = dialogRequest
			case "INSPECT":
				m.workspace.mode = modeInspect
			case "COMPARE":
				m.workspace.mode = modeCompare
			case "QUIT":
				m.workspace.dialog = dialogConfirmQuit
			}

			out := m.renderStatusline(m.ShellState(), width)

			if strings.Contains(out, "\n") {
				t.Fatalf("orient %q width %d: footer must not wrap", orient, width)
			}
			if w := lipgloss.Width(out); w != width {
				t.Fatalf("orient %q width %d: rendered width %d != terminal width", orient, width, w)
			}
			if ribbonDividerColumn(out) != ribbonBadgeWidth {
				t.Fatalf("orient %q width %d: divider not at fixed column %d", orient, width, ribbonBadgeWidth)
			}
			if n := strings.Count(stripANSI(out), "│"); n != 1 {
				t.Fatalf("orient %q width %d: expected exactly 1 divider, found %d", orient, width, n)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Badge regression
// ---------------------------------------------------------------------------

func TestFooter_BadgeRegressionAllWorkspaces(t *testing.T) {
	// Every workspace badge renders inside the same allocated region.
	labels := []string{"READY", "OBSERVE", "REQUEST", "INSPECT", "COMPARE", "QUIT", "EXECUTE"}

	for _, label := range labels {
		badge := renderWorkspaceBadge(label)
		if w := lipgloss.Width(badge); w != ribbonBadgeWidth {
			t.Fatalf("label %q: badge occupies %d cells, want fixed %d", label, w, ribbonBadgeWidth)
		}
		if !strings.Contains(stripANSI(badge), label) {
			t.Fatalf("label %q: badge must contain its label text", label)
		}
	}
}

func TestFooter_BadgeHighlightFlushAgainstDivider(t *testing.T) {
	// The colored highlight cell must terminate exactly at the divider column.
	for _, label := range []string{"READY", "OBSERVE", "REQUEST"} {
		badge := renderWorkspaceBadge(label)
		if w := lipgloss.Width(badge); w != ribbonBadgeWidth {
			t.Fatalf("label %q: highlight width %d != %d", label, w, ribbonBadgeWidth)
		}
		// The divider in a full render sits immediately after the badge.
		m := NewModel()
		out := m.renderStatusline(m.ShellState(), 100)
		if ribbonDividerColumn(out) != ribbonBadgeWidth {
			t.Fatalf("label %q: divider not flush at column %d", label, ribbonBadgeWidth)
		}
	}
}

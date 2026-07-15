package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

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
					out := m.View().Content
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

					view := m.View().Content
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
			_ = m.View().Content // must not panic
		})
	}
}

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

	a := m.View().Content
	b := m.View().Content

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

			view := m.View().Content
			if view == "" {
				t.Fatal("View() returned empty string")
			}
		})
	}
}

func TestPayloadGeometry_WidthConsistency(t *testing.T) {
	// After resize, widget widths must be consistent with the computed geometry.
	// textarea.Width() excludes the line-number gutter (4 chars), so the
	// text area width plus gutter equals the requested BodyWidth.
	m := NewModel()
	m.shell.Resize(100, 30)
	m.syncPayloadGeometry(100)

	geo := calculatePayloadGeometry(100)
	bodyW := m.bodyInput.Width()
	if bodyW != geo.BodyWidth {
		t.Fatalf("bodyInput.Width() = %d, want %d (BodyWidth = %d)",
			bodyW, geo.BodyWidth, geo.BodyWidth)
	}
	if len(m.headers) > 0 {
		if m.headers[0].Key.Width() != geo.KeyWidth {
			t.Fatalf("header[0].Key.Width = %d, want %d", m.headers[0].Key.Width(), geo.KeyWidth)
		}
		if m.headers[0].Value.Width() != geo.ValueWidth {
			t.Fatalf("header[0].Value.Width = %d, want %d", m.headers[0].Value.Width(), geo.ValueWidth)
		}
	}
}

func TestPayloadGeometry_ResizeIncreasesWidth(t *testing.T) {
	m := NewModel()
	m.shell.Resize(80, 24)
	m.syncPayloadGeometry(80)
	narrow := calculatePayloadGeometry(80).BodyWidth

	m.syncPayloadGeometry(160)
	wide := calculatePayloadGeometry(160).BodyWidth

	if wide <= narrow {
		t.Fatalf("BodyWidth at 160 (%d) must be > BodyWidth at 80 (%d)", wide, narrow)
	}
}

func TestPayloadGeometry_MinimumWidth(t *testing.T) {
	// At very narrow widths, geometry should floor at minimums.
	geo := calculatePayloadGeometry(20)
	if geo.AvailableWidth < 20 {
		t.Fatal("AvailableWidth must be at least 20")
	}
	if geo.KeyWidth < 10 {
		t.Fatal("KeyWidth must be at least 10")
	}
	if geo.ValueWidth < 10 {
		t.Fatal("ValueWidth must be at least 10")
	}
	if geo.BodyWidth < 10 {
		t.Fatal("BodyWidth must be at least 10")
	}
	if geo.BodyHeight < 1 {
		t.Fatal("BodyHeight must be at least 1")
	}
}

func TestPayloadGeometry_FlooredAtNarrow(t *testing.T) {
	// Widths below minimum should be clamped.
	geo := calculatePayloadGeometry(5)
	if geo.AvailableWidth < 20 {
		t.Fatal("calculatePayloadGeometry should floor its input at 20")
	}
}

func TestPayloadGeometry_WorkspaceContentWidth(t *testing.T) {
	tt := []struct {
		shell  int
		height int
		want   int
	}{
		{80, 24, 76},
		{100, 24, 96},
		{120, 24, 116},
		{30, 24, 68},
		{10, 24, 68},
	}
	for _, tc := range tt {
		got := workspaceContentWidth(tc.shell, tc.height)
		if got != tc.want {
			t.Errorf("workspaceContentWidth(%d, %d) = %d, want %d", tc.shell, tc.height, got, tc.want)
		}
	}
}

func TestPayloadGeometry_PayloadContentWidth(t *testing.T) {
	tt := []struct {
		shell  int
		height int
		want   int
	}{
		// Below context threshold: same as workspace content width
		{80, 24, 76},
		{100, 24, 96},
		{120, 24, 116},
		// Above context threshold (>=140): primary area is narrower
		// cw=156, ctx=156/3=52 (>28), primary=156-52-1=103
		{160, 40, 103},
		// cw=196, ctx=196/3=65 (>28), primary=196-65-1=130
		{200, 40, 130},
		// cw=216, ctx=216/3=72 (>28), primary=216-72-1=143
		{220, 40, 143},
	}
	for _, tc := range tt {
		got := payloadContentWidth(tc.shell, tc.height)
		if got != tc.want {
			t.Errorf("payloadContentWidth(%d, %d) = %d, want %d", tc.shell, tc.height, got, tc.want)
		}
	}
}

func TestV098Regression_PayloadGeometryContextPanelAware(t *testing.T) {
	// When context panel is visible (shellWidth >= 140), dialog-open and
	// WindowSizeMsg must set payload geometry accounting for the context
	// panel (narrower than full workspace content width).

	// --- Dialog opened via 'e' at 160 wide (context panel visible) ---
	m := NewModel()
	m.shell.Resize(160, 24)
	m2, _ := m.Update(keyMsgRune('e'))
	m = m2.(Model)

	if m.workspace.dialog != dialogRequest {
		t.Fatal("'e' should open Request dialog")
	}

	bodyW := m.bodyInput.Width()
	expected := calculatePayloadGeometry(payloadContentWidth(160, 24)).BodyWidth
	if bodyW != expected {
		t.Fatalf("dialog-open: body width = %d, want %d (payloadContentWidth=%d)",
			bodyW, expected, payloadContentWidth(160, 24))
	}

	withoutPanel := calculatePayloadGeometry(workspaceContentWidth(160, 24)).BodyWidth
	if bodyW >= withoutPanel {
		t.Fatalf("dialog-open: body width %d should account for context panel (full workspace would give %d)",
			bodyW, withoutPanel)
	}

	// --- WindowSizeMsg at 160 wide (context panel visible) ---
	m3 := NewModel()
	m3.shell.Resize(160, 24)
	m4, _ := m3.Update(tea.WindowSizeMsg{Width: 160, Height: 24})
	m3 = m4.(Model)

	bodyW2 := m3.bodyInput.Width()
	if bodyW2 != expected {
		t.Fatalf("WindowSizeMsg: body width = %d, want %d", bodyW2, expected)
	}

	// --- Control: at 100 wide (no context panel), body uses full workspace ---
	m5 := NewModel()
	m5.shell.Resize(100, 24)
	m6, _ := m5.Update(keyMsgRune('e'))
	m5 = m6.(Model)

	bodyW3 := m5.bodyInput.Width()
	expected3 := calculatePayloadGeometry(payloadContentWidth(100, 24)).BodyWidth
	if bodyW3 != expected3 {
		t.Fatalf("narrow shell: body width = %d, want %d", bodyW3, expected3)
	}

	narrowFull := calculatePayloadGeometry(workspaceContentWidth(100, 24)).BodyWidth
	if bodyW3 != narrowFull {
		t.Fatalf("at narrow shell, body width %d should match full workspace width %d",
			bodyW3, narrowFull)
	}
}

package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestMethodColor(t *testing.T) {
	tt := []struct {
		method string
		color  string
	}{
		{"GET", colorCyan},
		{"POST", colorGreen},
		{"PUT", colorAmber},
		{"DELETE", colorRose},
		{"PATCH", colorFuchsia},
		{"HEAD", colorMuted},
		{"OPTIONS", colorMuted},
	}

	for _, tc := range tt {
		got := methodColor(tc.method)
		if got != tc.color {
			t.Errorf("methodColor(%q) = %q (expected %q)", tc.method, got, tc.color)
		}
	}
}

func TestMethodStyle(t *testing.T) {
	s := methodStyle("GET", false)
	rendered := s.Render("GET")
	if !strings.Contains(rendered, "GET") {
		t.Fatal("methodStyle should contain the method name")
	}

	sFocused := methodStyle("POST", true)
	renderedFocused := sFocused.Render("POST")
	if !strings.Contains(renderedFocused, "POST") {
		t.Fatal("focused methodStyle should contain the method name")
	}

	sDefault := methodStyle("HEAD", false)
	renderedDefault := sDefault.Render("HEAD")
	if !strings.Contains(renderedDefault, "HEAD") {
		t.Fatal("default methodStyle should contain HEAD")
	}
}

func TestControlStyle(t *testing.T) {
	unfocused := controlStyle(false)
	rendered := unfocused.Render("test")
	if !strings.Contains(rendered, "test") {
		t.Fatal("controlStyle should render content")
	}

	focused := controlStyle(true)
	renderedFocused := focused.Render("test")
	if !strings.Contains(renderedFocused, "test") {
		t.Fatal("focused controlStyle should render content")
	}
}

func TestRunButtonStyle(t *testing.T) {
	idle := runButtonStyle(false)
	rendered := idle.Render("RUN")
	if !strings.Contains(rendered, "RUN") {
		t.Fatal("idle run button should contain RUN")
	}

	running := runButtonStyle(true)
	renderedRunning := running.Render("CANCEL")
	if !strings.Contains(renderedRunning, "CANCEL") {
		t.Fatal("running button should contain CANCEL")
	}
}

func TestPillStyle(t *testing.T) {
	active := pillStyle(true)
	rendered := active.Render("Timeline")
	if !strings.Contains(rendered, "Timeline") {
		t.Fatal("active pill should contain 'Timeline'")
	}

	inactive := pillStyle(false)
	renderedInactive := inactive.Render("Live Logs")
	if !strings.Contains(renderedInactive, "Live Logs") {
		t.Fatal("inactive pill should contain 'Live Logs'")
	}
}

func TestRowStyle(t *testing.T) {
	selected := rowStyle(true)
	rendered := selected.Render("test")
	if !strings.Contains(rendered, "test") {
		t.Fatal("selected row should contain text")
	}

	normal := rowStyle(false)
	renderedNormal := normal.Render("test")
	if !strings.Contains(renderedNormal, "test") {
		t.Fatal("normal row should contain text")
	}
}

func TestErrorRowStyle(t *testing.T) {
	selected := errorRowStyle(true)
	rendered := selected.Render("error")
	if !strings.Contains(rendered, "error") {
		t.Fatal("selected error row should contain text")
	}

	normal := errorRowStyle(false)
	renderedNormal := normal.Render("error")
	if !strings.Contains(renderedNormal, "error") {
		t.Fatal("normal error row should contain text")
	}
}

func TestBodyStyle(t *testing.T) {
	focused := bodyStyle(true)
	rendered := focused.Render("body")
	if !strings.Contains(rendered, "body") {
		t.Fatal("focused bodyStyle should contain text")
	}

	unfocused := bodyStyle(false)
	renderedUnfocused := unfocused.Render("body")
	if !strings.Contains(renderedUnfocused, "body") {
		t.Fatal("unfocused bodyStyle should contain text")
	}
}

func TestInputStyle(t *testing.T) {
	focused := inputStyle(true)
	rendered := focused.Render("input")
	if !strings.Contains(rendered, "input") {
		t.Fatal("focused inputStyle should contain text")
	}

	unfocused := inputStyle(false)
	renderedUnfocused := unfocused.Render("input")
	if !strings.Contains(renderedUnfocused, "input") {
		t.Fatal("unfocused inputStyle should contain text")
	}
}

func TestStatusColor(t *testing.T) {
	tt := []struct {
		status int
		color  string
	}{
		{0, colorRose},
		{200, colorGreen},
		{301, colorCyan},
		{404, colorRose},
		{500, colorRose},
	}

	for _, tc := range tt {
		got := statusColor(tc.status)
		if got != tc.color {
			t.Errorf("statusColor(%d) = %q (expected %q)", tc.status, got, tc.color)
		}
	}
}

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"colorBg", colorBg, "#09090b"},
		{"colorPanel", colorPanel, "#111113"},
		{"colorElevated", colorElevated, "#1c1c1e"},
		{"colorBorder", colorBorder, "#3f3f46"},
		{"colorText", colorText, "#d4d4d8"},
		{"colorMuted", colorMuted, "#71717a"},
		{"colorCyan", colorCyan, "#22d3ee"},
		{"colorCyanStrong", colorCyanStrong, "#06b6d4"},
		{"colorGreen", colorGreen, "#34d399"},
		{"colorAmber", colorAmber, "#fbbf24"},
		{"colorRose", colorRose, "#fb7185"},
		{"colorFuchsia", colorFuchsia, "#e879f9"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestStatusDotGlow_VisualFeedback(t *testing.T) {
	if got := statusDotStyle.GetForeground(); got != lipgloss.Color(colorCyan) {
		t.Errorf("statusDotStyle foreground = %q, want %q", got, colorCyan)
	}
	if got := statusDotGlowStyle.GetForeground(); got != lipgloss.Color(colorCyanStrong) {
		t.Errorf("statusDotGlowStyle foreground = %q, want %q", got, colorCyanStrong)
	}
	if got := statusDotGlowStyle.GetBold(); !got {
		t.Error("statusDotGlowStyle should be bold")
	}
	if got := statusDotStyle.GetBold(); got {
		t.Error("statusDotStyle should not be bold")
	}
}

func TestControlStyle_FocusVisualFeedback(t *testing.T) {
	focused := controlStyle(true)
	unfocused := controlStyle(false)

	if got := focused.GetForeground(); got != lipgloss.Color(colorCyan) {
		t.Errorf("focused controlStyle text fg = %q, want %q", got, colorCyan)
	}
	if got := unfocused.GetForeground(); got != lipgloss.Color(colorText) {
		t.Errorf("unfocused controlStyle text fg = %q, want %q", got, colorText)
	}
}

func TestInputStyle_FocusVisualFeedback(t *testing.T) {
	focused := inputStyle(true)
	unfocused := inputStyle(false)

	if got := focused.GetForeground(); got != lipgloss.Color(colorCyan) {
		t.Errorf("focused inputStyle text fg = %q, want %q", got, colorCyan)
	}
	if got := unfocused.GetForeground(); got != lipgloss.Color("#e4e4e7") {
		t.Errorf("unfocused inputStyle text fg = %q, want %q", got, "#e4e4e7")
	}
}

func TestErrorRowStyle_SelectedVisualFeedback(t *testing.T) {
	selected := errorRowStyle(true)
	normal := errorRowStyle(false)

	if got := selected.GetForeground(); got != lipgloss.Color(colorRose) {
		t.Errorf("selected errorRowStyle fg = %q, want %q", got, colorRose)
	}
	if got := normal.GetForeground(); got != lipgloss.Color(colorRose) {
		t.Errorf("normal errorRowStyle fg = %q, want %q", got, colorRose)
	}
	if got := selected.GetBackground(); got == nil {
		t.Error("selected errorRowStyle should have background")
	}
	if got := selected.GetBold(); !got {
		t.Error("selected errorRowStyle should be bold")
	}
}

func TestPillStyle_ActiveVisualFeedback(t *testing.T) {
	active := pillStyle(true)
	inactive := pillStyle(false)

	if got := active.GetForeground(); got != lipgloss.Color("#ffffff") {
		t.Errorf("active pillStyle fg = %q, want %q", got, "#ffffff")
	}
	if got := active.GetBackground(); got != lipgloss.Color(colorBorder) {
		t.Errorf("active pillStyle bg = %q, want %q", got, colorBorder)
	}
	if got := active.GetBold(); !got {
		t.Error("active pillStyle should be bold")
	}
	if got := inactive.GetForeground(); got != lipgloss.Color(colorMuted) {
		t.Errorf("inactive pillStyle fg = %q, want %q", got, colorMuted)
	}
	if got := inactive.GetBold(); got {
		t.Error("inactive pillStyle should not be bold")
	}
}

func TestMethodStyle_AppliesMethodColor(t *testing.T) {
	sGet := methodStyle("GET", false)
	if got := sGet.GetForeground(); got != lipgloss.Color(colorCyan) {
		t.Errorf("GET methodStyle fg = %q, want %q", got, colorCyan)
	}
	sPost := methodStyle("POST", false)
	if got := sPost.GetForeground(); got != lipgloss.Color(colorGreen) {
		t.Errorf("POST methodStyle fg = %q, want %q", got, colorGreen)
	}
	sDelete := methodStyle("DELETE", false)
	if got := sDelete.GetForeground(); got != lipgloss.Color(colorRose) {
		t.Errorf("DELETE methodStyle fg = %q, want %q", got, colorRose)
	}
}

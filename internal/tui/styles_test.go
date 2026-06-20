package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"colorBg", colorBg, "#09090b"},
		{"colorText", colorText, "#e4e4e7"},
		{"colorMuted", colorMuted, "#a1a1aa"},
		{"colorDark", colorDark, "#27272a"},
		{"colorAccent", colorAccent, "#38bdf8"},
		{"colorSuccess", colorSuccess, "#34d399"},
		{"colorWarning", colorWarning, "#fbbf24"},
		{"colorError", colorError, "#f87171"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestStyleBase(t *testing.T) {
	if got := StyleBase.GetForeground(); got != lipgloss.Color(colorText) {
		t.Errorf("StyleBase foreground = %q, want %q", got, colorText)
	}
	if got := StyleBase.GetBackground(); got != lipgloss.Color(colorBg) {
		t.Errorf("StyleBase background = %q, want %q", got, colorBg)
	}
}

func TestStyleStatusBar(t *testing.T) {
	if got := StyleStatusBar.GetForeground(); got != lipgloss.Color(colorText) {
		t.Errorf("StyleStatusBar foreground = %q, want %q", got, colorText)
	}
	if got := StyleStatusBar.GetBackground(); got != lipgloss.Color(colorDark) {
		t.Errorf("StyleStatusBar background = %q, want %q", got, colorDark)
	}
}

func TestStyleStatusMode(t *testing.T) {
	if got := StyleStatusMode.GetForeground(); got != lipgloss.Color(colorBg) {
		t.Errorf("StyleStatusMode foreground = %q, want %q", got, colorBg)
	}
	if got := StyleStatusMode.GetBackground(); got != lipgloss.Color(colorAccent) {
		t.Errorf("StyleStatusMode background = %q, want %q", got, colorAccent)
	}
	if got := StyleStatusMode.GetBold(); !got {
		t.Error("StyleStatusMode should be bold")
	}
}

func TestRowStyle(t *testing.T) {
	selected := rowStyle(true)
	rendered := selected.Render("test")
	if !strings.Contains(rendered, "test") {
		t.Fatal("selected row should contain text")
	}
	if got := selected.GetForeground(); got != lipgloss.Color(colorAccent) {
		t.Errorf("selected rowStyle fg = %q, want %q", got, colorAccent)
	}
	if got := selected.GetBackground(); got != lipgloss.Color(colorDark) {
		t.Errorf("selected rowStyle bg = %q, want %q", got, colorDark)
	}

	normal := rowStyle(false)
	renderedNormal := normal.Render("test")
	if !strings.Contains(renderedNormal, "test") {
		t.Fatal("normal row should contain text")
	}
	if got := normal.GetForeground(); got != lipgloss.Color(colorText) {
		t.Errorf("normal rowStyle fg = %q, want %q", got, colorText)
	}
}

func TestErrorRowStyle(t *testing.T) {
	selected := errorRowStyle(true)
	normal := errorRowStyle(false)

	if got := selected.GetForeground(); got != lipgloss.Color(colorError) {
		t.Errorf("selected errorRowStyle fg = %q, want %q", got, colorError)
	}
	if got := normal.GetForeground(); got != lipgloss.Color(colorError) {
		t.Errorf("normal errorRowStyle fg = %q, want %q", got, colorError)
	}
	if got := selected.GetBackground(); got == nil {
		t.Error("selected errorRowStyle should have background")
	}
	if got := selected.GetBold(); !got {
		t.Error("selected errorRowStyle should be bold")
	}
}

func TestStatusColor(t *testing.T) {
	tt := []struct {
		status int
		color  string
	}{
		{0, colorError},
		{200, colorSuccess},
		{301, colorWarning},
		{404, colorError},
		{500, colorError},
		{99, colorMuted},
	}

	for _, tc := range tt {
		got := statusColor(tc.status)
		if got != tc.color {
			t.Errorf("statusColor(%d) = %q (expected %q)", tc.status, got, tc.color)
		}
	}
}

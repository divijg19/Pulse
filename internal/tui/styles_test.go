package tui

import (
	"strings"
	"testing"
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

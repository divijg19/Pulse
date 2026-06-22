package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestV080Audit_AllSurfaces(t *testing.T) {
	for _, c := range AllAuditSurfaces() {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			for _, s := range AuditSizes {
				s := s
				t.Run(fmt.Sprintf("%dx%d", s.W, s.H), func(t *testing.T) {
					m := c.Setup()
					m.width = s.W
					m.height = s.H
					out := m.View()
					t.Logf("=== %s at %dx%d ===\n%s", c.Name, s.W, s.H, out)
					checkE1(t, out, s.W, s.H, c.Name)
				})
			}
		})
	}
}

func TestV080Audit_E1VerticalPadding(t *testing.T) {
	t.Run("Ready_80x24", func(t *testing.T) {
		m := newReadyModel()
		m.width = 80
		m.height = 24
		out := m.View()
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != 24 {
			t.Errorf("expected 24 lines, got %d", len(lines))
		}
		for i, line := range lines {
			vw := lipgloss.Width(line)
			if vw != 80 {
				t.Errorf("line %d: visual width %d, expected 80", i, vw)
			}
		}
	})
}

func TestV080Audit_E1StyleBleed(t *testing.T) {
	m := newTimelineRunningModel()
	for _, s := range AuditSizes {
		m.width = s.W
		m.height = s.H
		out := m.View()
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) != s.H {
			t.Errorf("%dx%d: expected %d lines, got %d", s.W, s.H, s.H, len(lines))
		}
		for i, line := range lines {
			vw := lipgloss.Width(line)
			if vw != s.W {
				t.Errorf("%dx%d line %d: visual width %d, expected %d", s.W, s.H, i, vw, s.W)
			}
		}
	}
}

func checkE1(t *testing.T, out string, width, height int, name string) {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != height {
		t.Errorf("%s %dx%d: line count %d, expected %d", name, width, height, len(lines), height)
	}
	for i, line := range lines {
		vw := lipgloss.Width(line)
		if vw != width {
			t.Errorf("%s %dx%d line %d: visual width %d, expected %d", name, width, height, i, vw, width)
		}
	}
}

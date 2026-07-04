package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRegionBoxed_RawWidthsMatch(t *testing.T) {
	// The raw (pre-lipgloss) border and content lines must have identical
	// character width. The top corner ┐ and bottom corner ┘ must appear at
	// the same column as the right wall │ of content lines.
	widths := []int{20, 40, 72}
	paddings := []int{0, 1, 2}

	for _, w := range widths {
		for _, p := range paddings {
			name := fmt.Sprintf("width=%d padding=%d", w, p)
			t.Run(name, func(t *testing.T) {
				r := Region{
					Width:    w,
					Height:   5,
					Border:   BorderFull,
					PaddingX: p,
				}

				rendered := r.renderBoxed("content")
				lines := strings.Split(rendered, "\n")

				if len(lines) < 3 {
					t.Fatal("expected >=3 lines")
				}

				topLine := lines[0]
				contentLine := lines[1]
				bottomLine := lines[len(lines)-1]

				// Check corners are at the end of their lines.
				topRunes := []rune(topLine)
				contentRunes := []rune(contentLine)
				bottomRunes := []rune(bottomLine)

				if string(topRunes[len(topRunes)-1]) != "┐" {
					t.Errorf("top border last char = %q at rune len %d, want ┐",
						string(topRunes[len(topRunes)-1]), len(topRunes))
				}
				if string(contentRunes[len(contentRunes)-1]) != "│" {
					t.Errorf("content line last char = %q at rune len %d, want │",
						string(contentRunes[len(contentRunes)-1]), len(contentRunes))
				}
				if string(bottomRunes[len(bottomRunes)-1]) != "┘" {
					t.Errorf("bottom border last char = %q at rune len %d, want ┘",
						string(bottomRunes[len(bottomRunes)-1]), len(bottomRunes))
				}

				// All three must have same rune count.
				if len(topRunes) != len(contentRunes) || len(contentRunes) != len(bottomRunes) {
					t.Errorf("raw rune lengths differ: top=%d content=%d bottom=%d (padding=%d)",
						len(topRunes), len(contentRunes), len(bottomRunes), p)
				}

				// Also verify consistent lipgloss width.
				tw := lipgloss.Width(topLine)
				cw := lipgloss.Width(contentLine)
				bw := lipgloss.Width(bottomLine)
				if tw != cw || cw != bw {
					t.Errorf("lipgloss widths differ: top=%d content=%d bottom=%d",
						tw, cw, bw)
				}
			})
		}
	}
}

func TestRegionBoxed_TitledCorners(t *testing.T) {
	r := Region{
		Width:    40,
		Height:   5,
		Border:   BorderFull,
		PaddingX: 1,
		Title:    "OBSERVE",
	}

	rendered := r.renderBoxed("content")
	lines := strings.Split(rendered, "\n")

	topLine := lines[0]
	contentLine := lines[1]
	bottomLine := lines[len(lines)-1]

	topRunes := []rune(topLine)
	contentRunes := []rune(contentLine)
	bottomRunes := []rune(bottomLine)

	if string(topRunes[len(topRunes)-1]) != "┐" {
		t.Errorf("titled top corner = %q, want ┐", string(topRunes[len(topRunes)-1]))
	}
	if string(contentRunes[len(contentRunes)-1]) != "│" {
		t.Errorf("titled content wall = %q, want │", string(contentRunes[len(contentRunes)-1]))
	}
	if string(bottomRunes[len(bottomRunes)-1]) != "┘" {
		t.Errorf("titled bottom corner = %q, want ┘", string(bottomRunes[len(bottomRunes)-1]))
	}

	if len(topRunes) != len(contentRunes) || len(contentRunes) != len(bottomRunes) {
		t.Errorf("titled rune lengths differ: top=%d content=%d bottom=%d",
			len(topRunes), len(contentRunes), len(bottomRunes))
	}
}

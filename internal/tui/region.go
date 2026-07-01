package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RegionType distinguishes the three kinds of region in the Shell layout.
type RegionType int

const (
	RegionUnset RegionType = iota
	ContextRegion
	WorkspaceRegion
	CommandRegion
)

// BorderStyle defines optional border decoration for a Region.
type BorderStyle int

const (
	BorderNone BorderStyle = iota // no border
	BorderFull                    // full box ┌┐└┘│─
)

// Region is a rectangular area within a layout. Each Region has a type that
// identifies its role in the composition. Regions may optionally carry
// presentation hints (border, title, padding) for the rendering layer.
type Region struct {
	Type    RegionType
	Width   int
	Height  int
	Border  BorderStyle
	Title   string
	Padding int  // horizontal padding inside any border
	NoClip  bool // allow content to specify its own height
}

// RenderBordered wraps content inside the region's border (if any).
func (r Region) RenderBordered(content string) string {
	if r.Border == BorderNone {
		return r.renderPadded(content)
	}
	return r.renderBoxed(content)
}

func (r Region) renderPadded(content string) string {
	if r.Padding <= 0 {
		return r.styleBase().Render(content)
	}
	pad := strings.Repeat(" ", r.Padding)
	lines := strings.Split(content, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = pad + lines[i]
		}
	}
	return r.styleBase().Render(strings.Join(lines, "\n"))
}

func (r Region) renderBoxed(content string) string {
	borderPad := r.Padding
	if borderPad < 0 {
		borderPad = 0
	}
	innerW := r.Width - 2 - 2*borderPad
	if innerW < 1 {
		innerW = 1
	}
	innerH := r.Height - 2
	if innerH < 0 {
		innerH = 0
	}
	pad := strings.Repeat(" ", borderPad)

	var b strings.Builder

	// Top border: title centered inside ┌──┐, or plain ┌─────┐
	topRule := strings.Repeat("─", innerW)
	if r.Title != "" {
		title := " " + r.Title + " "
		titleLen := lipgloss.Width(title)
		if titleLen > innerW {
			title = title[:innerW]
			titleLen = innerW
		}
		leftCount := (innerW - titleLen) / 2
		rightCount := innerW - titleLen - leftCount
		leftRule := strings.Repeat("─", leftCount)
		rightRule := strings.Repeat("─", rightCount)
		b.WriteString("┌" + leftRule + title + rightRule + "┐\n")
	} else {
		b.WriteString("┌" + topRule + "┐\n")
	}

	// Content lines: │ pad content padspaces pad │
	contentLines := strings.Split(content, "\n")
	for i := 0; i < innerH; i++ {
		if i < len(contentLines) {
			line := contentLines[i]
			padWidth := innerW - lipgloss.Width(line)
			if padWidth < 0 {
				padWidth = 0
			}
			b.WriteString("│" + pad + line + strings.Repeat(" ", padWidth) + pad + "│\n")
		} else {
			b.WriteString("│" + strings.Repeat(" ", innerW+2*borderPad) + "│\n")
		}
	}

	// Bottom border: └─────┘
	b.WriteString("└" + strings.Repeat("─", innerW) + "┘\n")

	return r.styleBase().Render(strings.TrimSuffix(b.String(), "\n"))
}

func (r Region) styleBase() lipgloss.Style {
	return lipgloss.NewStyle().Width(r.Width).Height(r.Height)
}

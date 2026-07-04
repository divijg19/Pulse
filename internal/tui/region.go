package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RegionType distinguishes the three kinds of region in the Shell layout.
type RegionType int

const (
	ContextRegion RegionType = iota
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
	Type     RegionType
	Width    int
	Height   int
	Border   BorderStyle
	Title    string
	PaddingX int // horizontal padding inside any border
	PaddingY int // vertical padding inside any border
}

// RenderBordered wraps content inside the region's border (if any).
func (r Region) RenderBordered(content string) string {
	if r.Border == BorderNone {
		return r.renderPadded(content)
	}
	return r.renderBoxed(content)
}

func (r Region) renderPadded(content string) string {
	px := r.paddingX()
	if px == 0 {
		return r.styleBase().Render(content)
	}
	pad := strings.Repeat(" ", px)
	lines := strings.Split(content, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = pad + lines[i]
		}
	}
	return r.styleBase().Render(strings.Join(lines, "\n"))
}

func (r Region) renderBoxed(content string) string {
	borderPadX := r.paddingX()
	paddingY := r.effectivePaddingY()
	innerW := r.Width - 2 - 2*borderPadX
	if innerW < 1 {
		innerW = 1
	}
	innerH := r.Height - 2
	if innerH < 0 {
		innerH = 0
	}
	pad := strings.Repeat(" ", borderPadX)

	var b strings.Builder

	borderDashWidth := innerW + 2*borderPadX
	if borderDashWidth < 0 {
		borderDashWidth = 0
	}
	topRule := strings.Repeat("─", borderDashWidth)
	if r.Title != "" {
		title := " " + r.Title + " "
		titleLen := lipgloss.Width(title)
		if titleLen > borderDashWidth {
			title = title[:borderDashWidth]
			titleLen = borderDashWidth
		}
		leftCount := (borderDashWidth - titleLen) / 2
		rightCount := borderDashWidth - titleLen - leftCount
		leftRule := strings.Repeat("─", leftCount)
		rightRule := strings.Repeat("─", rightCount)
		b.WriteString("┌")
		b.WriteString(leftRule)
		b.WriteString(title)
		b.WriteString(rightRule)
		b.WriteString("┐\n")
	} else {
		b.WriteString("┌")
		b.WriteString(topRule)
		b.WriteString("┐\n")
	}

	// Content lines: │ pad content padspaces pad │
	// Vertical padding (PaddingY) adds empty rows at the top and bottom
	// of the bordered area.
	contentLines := strings.Split(content, "\n")
	for i := 0; i < innerH; i++ {
		if i < paddingY || i >= paddingY+len(contentLines) {
			b.WriteString("│")
			b.WriteString(strings.Repeat(" ", innerW+2*borderPadX))
			b.WriteString("│\n")
		} else {
			line := contentLines[i-paddingY]
			padWidth := innerW - lipgloss.Width(line)
			if padWidth < 0 {
				padWidth = 0
			}
			b.WriteString("│" + pad + line)
			b.WriteString(strings.Repeat(" ", padWidth))
			b.WriteString(pad + "│\n")
		}
	}

	// Bottom border: └─────┘ (same width as top border)
	b.WriteString("└")
	b.WriteString(strings.Repeat("─", borderDashWidth))
	b.WriteString("┘\n")

	return r.styleBase().Render(strings.TrimSuffix(b.String(), "\n"))
}

// paddingX returns the effective horizontal padding, normalizing
// negative values to zero.
func (r Region) paddingX() int {
	if r.PaddingX < 0 {
		return 0
	}
	return r.PaddingX
}

// paddingY returns the effective vertical padding, normalizing
// negative values to zero.
func (r Region) paddingY() int {
	if r.PaddingY < 0 {
		return 0
	}
	return r.PaddingY
}

// minimumContentLines is the smallest number of content lines
// guaranteed inside a bordered region regardless of padding. This
// prevents padding from consuming all available space.
const minimumContentLines = 3

// effectivePaddingY returns the vertical padding to apply inside a
// bordered region. The value may be reduced on constrained heights
// to reserve at least 3 content lines. The policy is owned by Region
// so that callers never need to reason about when padding applies.
func (r Region) effectivePaddingY() int {
	py := r.paddingY()
	if py == 0 {
		return 0
	}
	if r.Border != BorderNone {
		innerH := r.Height - 2
		// Reserve at least minimumContentLines before applying padding.
		maxPad := (innerH - minimumContentLines) / 2
		if maxPad < 1 {
			return 0
		}
		if py > maxPad {
			return maxPad
		}
	}
	return py
}

// ContentRegion returns the content viewport — the region available
// for surfaces to render into after borders and padding are subtracted.
// The returned Region carries the caller's Type so surfaces
// receive semantic context alongside dimensions.
func (r Region) ContentRegion() Region {
	inner := Region{
		Type: r.Type,
	}

	inner.Width = r.Width
	if r.Border != BorderNone {
		inner.Width -= 2
	}
	inner.Width -= 2 * r.paddingX()
	if inner.Width < 1 {
		inner.Width = 1
	}

	inner.Height = r.Height
	if r.Border != BorderNone {
		inner.Height -= 2
	}
	inner.Height -= 2 * r.effectivePaddingY()
	if inner.Height < 0 {
		inner.Height = 0
	}

	return inner
}

func (r Region) styleBase() lipgloss.Style {
	return lipgloss.NewStyle().Width(r.Width).Height(r.Height)
}

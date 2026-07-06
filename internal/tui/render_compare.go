package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const minBodyWidth = 10

func (m Model) renderCompare(region Region) string {
	if m.workspace.compare.marked < 0 || m.workspace.compare.active < 0 {
		return regionStyle(region).Render(styleMuted.Render("No comparison active."))
	}

	markedResult := m.results[m.workspace.compare.marked]
	activeResult := m.results[m.workspace.compare.active]

	markedMethod := m.effectiveMethod(markedResult)
	markedURL := m.effectiveURL(markedResult)
	activeMethod := m.effectiveMethod(activeResult)
	activeURL := m.effectiveURL(activeResult)

	markedIdentity := styleCompareMarked.Render("◆ Marked")
	activeIdentity := styleCompareActive.Render("▶ Current")

	markedWhat := m.renderInspectSummary(markedResult, markedMethod, markedURL)
	activeWhat := m.renderInspectSummary(activeResult, activeMethod, activeURL)

	halfH := max(3, region.Height/2-2)
	halfW := max(minBodyWidth, region.Width/2-2)

	markedWhy := m.renderInspectWhy(markedResult, halfH)
	activeWhy := m.renderInspectWhy(activeResult, halfH)

	markedBody := m.renderInspectBody(markedResult, halfH, halfW)
	activeBody := m.renderInspectBody(activeResult, halfH, halfW)

	switch {
	case region.Width >= 120:
		return m.renderCompareWide(region, markedIdentity, markedWhat, markedWhy, markedBody,
			activeIdentity, activeWhat, activeWhy, activeBody)
	case region.Width >= 80:
		return m.renderCompareMedium(region, markedIdentity, markedWhat, markedWhy, markedBody,
			activeIdentity, activeWhat, activeWhy, activeBody)
	default:
		return regionStyle(region).Render(styleMuted.Render("Comparison requires at least 80 columns."))
	}
}

func (m Model) renderCompareWide(region Region, markedId, markedWhat, markedWhy, markedBody string,
	activeId, activeWhat, activeWhy, activeBody string) string {

	whatW := max(minBodyWidth, region.Width*35/100)
	whyW := max(minBodyWidth, region.Width*25/100)
	bodyW := max(minBodyWidth, region.Width-whatW-whyW-2)

	markedWhatR := lipgloss.NewStyle().Width(whatW).Render(
		sectionLine("WHAT HAPPENED", false) + "\n" + markedWhat)
	markedWhyR := lipgloss.NewStyle().Width(whyW).Render(
		sectionLine("WHY", false) + "\n" + markedWhy)
	markedBodyR := lipgloss.NewStyle().Width(bodyW).Render(
		sectionLine("RESPONSE", false) + "\n" + markedBody)

	activeWhatR := lipgloss.NewStyle().Width(whatW).Render(
		sectionLine("WHAT HAPPENED", false) + "\n" + activeWhat)
	activeWhyR := lipgloss.NewStyle().Width(whyW).Render(
		sectionLine("WHY", false) + "\n" + activeWhy)
	activeBodyR := lipgloss.NewStyle().Width(bodyW).Render(
		sectionLine("RESPONSE", false) + "\n" + activeBody)

	markedCols := lipgloss.JoinHorizontal(lipgloss.Top, markedWhatR, " ", markedWhyR, " ", markedBodyR)
	activeCols := lipgloss.JoinHorizontal(lipgloss.Top, activeWhatR, " ", activeWhyR, " ", activeBodyR)

	markedPane := lipgloss.NewStyle().Render(markedId + gapSection + markedCols)
	activePane := lipgloss.NewStyle().Render(activeId + gapSection + activeCols)

	return regionStyle(region).Render(lipgloss.JoinHorizontal(lipgloss.Top, markedPane, " │ ", activePane))
}

func (m Model) renderCompareMedium(region Region, markedId, markedWhat, markedWhy, markedBody string,
	activeId, activeWhat, activeWhy, activeBody string) string {

	markedTitle := markedId + gapSection + sectionLine("WHAT HAPPENED", false) + "\n" + markedWhat
	activeTitle := activeId + gapSection + sectionLine("WHAT HAPPENED", false) + "\n" + activeWhat
	markedWhyTitle := sectionLine("WHY", false) + "\n" + markedWhy
	activeWhyTitle := sectionLine("WHY", false) + "\n" + activeWhy
	markedBodyTitle := sectionLine("RESPONSE", false) + "\n" + markedBody
	activeBodyTitle := sectionLine("RESPONSE", false) + "\n" + activeBody

	lines := []string{
		markedTitle,
		"\n",
		markedWhyTitle,
		"\n",
		markedBodyTitle,
		"\n",
		activeTitle,
		"\n",
		activeWhyTitle,
		"\n",
		activeBodyTitle,
	}

	return regionStyle(region).Render(strings.Join(lines, "\n"))
}

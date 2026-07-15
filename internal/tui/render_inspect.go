package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m Model) renderInspect(region Region) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return regionStyle(region).Render(styleMuted.Render("No result selected."))
	}

	result := m.results[m.selected]
	method := m.effectiveMethod(result)
	reqURL := m.effectiveURL(result)
	identity := identityCell(fmt.Sprintf("Result %d", m.selected+1))

	what := m.renderInspectSummary(result, method, reqURL)

	switch {
	case region.Width >= 90:
		whatW := region.Width * 35 / 100
		whyW := region.Width * 25 / 100
		bodyW := region.Width - whatW - whyW - 2
		why := m.renderInspectWhy(result, region.Height-2)
		body := m.renderInspectBody(result, region.Height-2, bodyW)

		whatRendered := lipgloss.NewStyle().Width(whatW).Render(
			sectionLine("WHAT HAPPENED", m.inspectZone == zoneWhatHappened) + "\n" + what)
		whyRendered := lipgloss.NewStyle().Width(whyW).Render(
			sectionLine("WHY", m.inspectZone == zoneWhy) + "\n" + why)
		bodyRendered := lipgloss.NewStyle().Width(bodyW).Render(
			sectionLine("RESPONSE", m.inspectZone == zoneBody) + "\n" + body)

		columns := lipgloss.JoinHorizontal(lipgloss.Top, whatRendered, " ", whyRendered, " ", bodyRendered)
		combined := identity + gapSection + columns
		return regionStyle(region).Render(combined)

	case region.Width >= 60:
		summaryH := strings.Count(what, "\n") + 4
		halfW := (region.Width - 1) / 2
		remainingH := region.Height - summaryH - 2
		bodyW := region.Width - halfW - 1

		top := lipgloss.NewStyle().Width(region.Width).Render(
			identity + gapSection + sectionLine("WHAT HAPPENED", m.inspectZone == zoneWhatHappened) + "\n" + what)

		whyRendered := lipgloss.NewStyle().Width(halfW).Render(
			sectionLine("WHY", m.inspectZone == zoneWhy) + "\n" + m.renderInspectWhy(result, remainingH-1))
		bodyRendered := lipgloss.NewStyle().Width(bodyW).Render(
			sectionLine("RESPONSE", m.inspectZone == zoneBody) + "\n" + m.renderInspectBody(result, remainingH-1, bodyW))

		bottom := lipgloss.JoinHorizontal(lipgloss.Top, whyRendered, " ", bodyRendered)
		combined := top + "\n" + bottom
		return regionStyle(region).Render(combined)

	default:
		whatLines := strings.Count(what, "\n") + 1

		// Fixed overhead before WHY content:
		//   identity(1) + gapSection(3) + sectionLine WHAT(1) +
		//   gapSection(3) + sectionLine WHY(1) = 9
		whyMaxH := region.Height - whatLines - 9
		if whyMaxH < 1 {
			whyMaxH = 1
		}

		whyRendered := m.renderInspectWhy(result, whyMaxH)
		whyLines := strings.Count(whyRendered, "\n") + 1

		// Fixed overhead before BODY content:
		//   what overhead (9) + gapSection(3) + sectionLine RESPONSE(1) = 13
		//   includes identity, all gapSections, and all sectionLines.
		bodyMaxH := region.Height - whatLines - whyLines - 13
		if bodyMaxH < 1 {
			bodyMaxH = 1
		}

		bodyText := m.renderInspectBody(result, bodyMaxH, region.Width)
		lines := []string{
			identity,
			gapSection,
			sectionLine("WHAT HAPPENED", m.inspectZone == zoneWhatHappened),
			what,
			gapSection,
			sectionLine("WHY", m.inspectZone == zoneWhy),
			whyRendered,
			gapSection,
			sectionLine("RESPONSE", m.inspectZone == zoneBody),
			bodyText,
		}

		return regionStyle(region).Render(strings.Join(lines, "\n"))
	}
}

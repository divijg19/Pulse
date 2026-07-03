package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/model"
)

func (m Model) renderInspectWhatHappened(result model.Result, method, reqURL string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s", renderMethod(method), reqURL))
	b.WriteString("\n")
	b.WriteString(renderStatusBadge(result))
	b.WriteString("\n")
	b.WriteString(renderMetadata("Latency", formatLatency(result.Latency)))
	return b.String()
}

func (m Model) renderInspectWhy(result model.Result, maxLines int) string {
	var b strings.Builder
	if result.Error != "" {
		b.WriteString(styleError.Render("Error: " + result.Error))
		return b.String()
	}
	if len(result.ResponseHeaders) == 0 {
		b.WriteString(styleMuted.Render("No headers captured."))
	} else {
		keys := make([]string, 0, len(result.ResponseHeaders))
		for key := range result.ResponseHeaders {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := 0
		for _, key := range keys {
			if maxLines > 0 && lines >= maxLines {
				b.WriteString(styleMuted.Render("..."))
				break
			}
			b.WriteString(renderMetadata(key, result.ResponseHeaders[key]))
			b.WriteString("\n")
			lines++
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m Model) renderInspectBodyText(result model.Result, maxLines int) string {
	return renderBodyPreview(result.ResponseBody, maxLines)
}

func (m Model) renderInspect(region Region) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return regionStyle(region).Render(styleMuted.Render("No result selected."))
	}

	result := m.results[m.selected]
	method := m.effectiveMethod(result)
	reqURL := m.effectiveURL(result)
	identity := identityCell(fmt.Sprintf("Result %d · Investigate", m.selected+1))

	sectionLine := func(label string) string {
		return styleSectionLine.Render("── " + label + " ──")
	}

	what := m.renderInspectWhatHappened(result, method, reqURL)
	why := m.renderInspectWhy(result, region.Height-2)
	body := m.renderInspectBodyText(result, region.Height-2)

	switch {
	case region.Width >= 90:
		whatW := region.Width * 28 / 100
		whyW := region.Width * 34 / 100
		bodyW := region.Width - whatW - whyW - 2

		whatRendered := lipgloss.NewStyle().Width(whatW).Render(
			identity + "\n\n" + sectionLine("WHAT HAPPENED") + "\n" + what)
		whyRendered := lipgloss.NewStyle().Width(whyW).Render(
			sectionLine("WHY") + "\n" + why)
		bodyRendered := lipgloss.NewStyle().Width(bodyW).Render(
			sectionLine("EXACTLY WHAT CAME BACK") + "\n" + body)

		combined := lipgloss.JoinHorizontal(lipgloss.Top, whatRendered, " ", whyRendered, " ", bodyRendered)
		return regionStyle(region).Render(combined)

	case region.Width >= 60:
		summaryH := 7
		halfW := (region.Width - 1) / 2
		remainingH := region.Height - summaryH - 2

		top := lipgloss.NewStyle().Width(region.Width).Render(
			identity + "\n\n" + sectionLine("WHAT HAPPENED") + "\n" + what)

		whyRendered := lipgloss.NewStyle().Width(halfW).Render(
			sectionLine("WHY") + "\n" + m.renderInspectWhy(result, remainingH-1))
		bodyRendered := lipgloss.NewStyle().Width(region.Width - halfW - 1).Render(
			sectionLine("EXACTLY WHAT CAME BACK") + "\n" + m.renderInspectBodyText(result, remainingH-1))

		bottom := lipgloss.JoinHorizontal(lipgloss.Top, whyRendered, " ", bodyRendered)
		combined := top + "\n" + bottom
		return regionStyle(region).Render(combined)

	default:
		lines := []string{
			identity,
			gapSection,
			sectionLine("WHAT HAPPENED"),
			what,
			gapSection,
			sectionLine("WHY"),
			m.renderInspectWhy(result, region.Height-8),
			gapSection,
			sectionLine("EXACTLY WHAT CAME BACK"),
			m.renderInspectBodyText(result, region.Height-10),
		}

		return regionStyle(region).Render(strings.Join(lines, "\n"))
	}
}

package tui

import (
	"fmt"
	"strings"

	"github.com/divijg19/Pulse/internal/runconfig"
)

func (m Model) renderContextRegion(region Region) string {
	switch {
	case m.workspace.dialog == dialogRequest:
		return m.renderRequestContext(region)
	case m.workspace.mode == modeCompare:
		return "" // Compare workspace is itself the main region
	case m.workspace.mode == modeInspect && len(m.results) > 0:
		return m.renderInspectContext(region)
	case m.workspace.compare.HasBaseline() || m.workspace.compare.HasReference():
		return m.renderComparePreview(region)
	case len(m.results) > 0:
		return m.renderObserveContext(region)
	default:
		return ""
	}
}

func (m Model) renderComparePreview(region Region) string {
	w := m.workspace.compare
	var b strings.Builder
	b.WriteString(accentOrMuted("Comparison", true))
	b.WriteString(gapSection)

	// Identity: baseline/candidate when a workspace is active, otherwise the
	// reference request (persisted across runs) on its own.
	if w.Baseline != nil || w.Candidate != nil {
		b.WriteString(m.renderComparisonIdentityBlock(w.Baseline, w.Candidate))
	} else if w.Reference != nil {
		b.WriteString(m.renderIdentityLine("●", "Pinned Baseline", *w.Reference))
	}

	if w.IsComparing() && w.Analysis != nil {
		b.WriteString("\n\n")
		b.WriteString(renderVerdict(w.Analysis))
	}

	// Context-specific keybindings make the next action obvious without leaving
	// the drawer, including how to renounce a reference.
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render(w.KeybindingHints()))

	return regionStyle(region).Render(b.String())
}

func (m Model) renderObserveContext(region Region) string {
	sel := m.selected
	if sel < 0 || sel >= len(m.results) {
		return regionStyle(region).Render("")
	}
	result := m.results[sel]
	method := m.effectiveMethod(result)
	reqURL := m.effectiveURL(result)

	var b strings.Builder
	b.WriteString(accentOrMuted("Selection", true))
	b.WriteString(gapSection)
	b.WriteString(fmt.Sprintf(indentField+"%s %s\n", method, truncateURL(reqURL, region.Width-contextRowWidth)))
	b.WriteString(fmt.Sprintf(indentField+"%s\n", renderStatusBadge(result)))
	b.WriteString(fmt.Sprintf(indentField+"Latency: %s\n", formatDuration(result.Latency)))
	return regionStyle(region).Render(b.String())
}

func (m Model) renderInspectContext(region Region) string {
	var b strings.Builder
	b.WriteString(accentOrMuted("Run Metrics", false))
	b.WriteString(gapSection)
	b.WriteString(fmt.Sprintf(indentField+"Duration: %s\n", formatDuration(m.elapsed)))
	b.WriteString(fmt.Sprintf(indentField+"Requests: %d\n", len(m.results)))
	if metrics := m.metricsString(); metrics != "" {
		b.WriteString(fmt.Sprintf(indentField+"%s\n", metrics))
	}
	return regionStyle(region).Render(b.String())
}

func (m Model) renderRequestContext(region Region) string {
	var b strings.Builder
	b.WriteString(accentOrMuted("Configuration", false))
	b.WriteString(gapSection)
	b.WriteString(fmt.Sprintf(indentField+"Method: %s\n", runconfig.AllowedMethods()[m.methodIndex]))
	b.WriteString(fmt.Sprintf(indentField+"URL: %s\n", truncateURL(m.urlInput.Value(), region.Width-contextURLWidth)))
	b.WriteString(fmt.Sprintf(indentField+"CC: %d\n", m.concurrency()))
	b.WriteString(fmt.Sprintf(indentField+"Payload: %s\n", m.payloadSummary()))
	return regionStyle(region).Render(b.String())
}

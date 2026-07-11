package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

func (m Model) renderResultList(region Region, identity string, emptyRunning string, rowFn func(result model.Result, index int, selected bool, width int) string) string {
	var b strings.Builder

	b.WriteString(identityCell(identity))
	b.WriteString(gapSection)

	remaining := region.Height - 2
	if metrics := m.metricsString(); metrics != "" {
		b.WriteString(styleMuted.Width(region.Width).Render(metrics))
		b.WriteString("\n")
		remaining--
	}

	if remaining <= 0 {
		return regionStyle(region).Render(b.String())
	}

	if len(m.results) == 0 {
		msg := m.renderEmptyState(emptyRunning)
		b.WriteString(styleMuted.Render(msg))
		return regionStyle(region).Render(b.String())
	}

	start := visibleWindow(len(m.results), m.selected, remaining)
	rows := make([]string, 0, min(len(m.results)-start, remaining))
	for i := start; i < len(m.results) && len(rows) < remaining; i++ {
		result := m.results[i]
		sel := i == m.selected
		rows = append(rows, rowFn(result, i, sel, region.Width))
	}
	b.WriteString(strings.Join(rows, "\n"))

	return regionStyle(region).Render(b.String())
}

func (m Model) renderTimeline(region Region) string {
	return m.renderResultList(region, "Timeline · Completed", "Waiting for completions...",
		func(result model.Result, index int, selected bool, width int) string {
			return m.renderTimelineRow(index, result, m.summary.MaxLatency, width, selected)
		})
}

func (m Model) renderTimelineRow(index int, result model.Result, maxLatency time.Duration, width int, selected bool) string {
	status := resultStatus(result)
	latency := formatDuration(result.Latency)
	method := m.effectiveMethod(result)
	reqURL := m.effectiveURL(result)

	remaining := width - timelineFixedWidth
	barWidth := max(6, remaining/2)
	urlWidth := max(10, remaining-barWidth)

	filled := 0
	if maxLatency > 0 {
		filled = int(float64(result.Latency) / float64(maxLatency) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
	}

	barColor := statusColor(result.Status)
	bar := renderLatencyBar(filled, barWidth, barColor)

	prefix := m.resultCompareMarker(index)

	line := fmt.Sprintf("%s%s %-12s %-4s %s %s %s",
		rowCursor(selected), prefix, status, method, bar, latency, truncateURL(reqURL, urlWidth))

	return renderResultRow(line, result, selected, width)
}

func (m Model) renderLogs(region Region) string {
	return m.renderResultList(region, "Logs · Sequence", "No events captured yet...",
		func(result model.Result, index int, selected bool, width int) string {
			stamp := requestTime(result)
			method := m.effectiveMethod(result)
			reqURL := m.effectiveURL(result)

			prefix := m.resultCompareMarker(index)

			line := fmt.Sprintf("%s%s #%03d %s %-4s %-10s %s %s",
				rowCursor(selected), prefix, index+1, stamp, method, resultStatus(result), formatDuration(result.Latency), truncate(reqURL, width-logsFixedWidth-logsFixedSuffix))
			if result.Error != "" {
				line = fmt.Sprintf("%s · %s", line, result.Error)
			}
			return renderResultRow(line, result, selected, width)
		})
}

// resultCompareMarker returns the timeline marker for the result at index,
// applying deterministic priority: Candidate > Baseline > Reference. Exactly one
// marker is returned per entry.
func (m Model) resultCompareMarker(index int) string {
	w := m.workspace.compare
	if index < 0 || index >= len(m.results) {
		return ""
	}
	switch {
	case w.Candidate != nil && resultsEqual(*w.Candidate, m.results[index]):
		return "▶ "
	case w.Baseline != nil && resultsEqual(*w.Baseline, m.results[index]):
		return "◆ "
	case w.Reference != nil && resultsEqual(*w.Reference, m.results[index]):
		return "● "
	}
	return ""
}

func (m Model) renderEmptyState(runningMsg string) string {
	if !m.running {
		if strings.TrimSpace(m.urlInput.Value()) == "" {
			return "Enter a URL to begin"
		}
		return "Ready"
	}
	return runningMsg
}

func renderResultRow(line string, result model.Result, selected bool, width int) string {
	if isErrorResult(result) {
		return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
	}
	return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
}

func resultsEqual(a, b model.Result) bool {
	return a.Status == b.Status &&
		a.Latency == b.Latency &&
		a.Error == b.Error &&
		a.RequestMethod == b.RequestMethod &&
		a.RequestURL == b.RequestURL &&
		a.ResponseBody == b.ResponseBody
}

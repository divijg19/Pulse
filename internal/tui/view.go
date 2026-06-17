package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/metrics"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
)

var sparklineChars = []rune("▁▂▃▄▅▆▇█")

func (m Model) View() string {
	if m.width == 0 {
		return "Pulse is starting..."
	}

	width := max(72, m.width)
	innerWidth := width - 4

	header := m.renderHeader(innerWidth)
	command := m.renderCommand(innerWidth)
	metricsStrip := m.renderMetrics(innerWidth)
	sparkline := m.renderSparkline(innerWidth)

	payload := ""
	if m.showPayload {
		payload = m.renderPayload(innerWidth)
	}

	fixed := lipgloss.Height(header)
	fixed += lipgloss.Height(command)
	fixed += lipgloss.Height(metricsStrip)
	if sparkline != "" {
		fixed += lipgloss.Height(sparkline)
	}
	if payload != "" {
		fixed += lipgloss.Height(payload)
	}

	footer := m.renderFooter(innerWidth)
	footerHeight := lipgloss.Height(footer)

	available := m.height - 2 - fixed - footerHeight - 4
	workspaceHeight := max(8, available)
	workspace := m.renderWorkspace(innerWidth, workspaceHeight)

	parts := []string{header, command, metricsStrip}
	if sparkline != "" {
		parts = append(parts, sparkline)
	}
	if payload != "" {
		parts = append(parts, payload)
	}
	parts = append(parts, workspace, footer)

	return appStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderHeader(width int) string {
	rightText := m.status
	if m.running {
		rps := 0.0
		if m.elapsed > 0 {
			rps = float64(len(m.results)) / m.elapsed.Seconds()
		}
		rightText = fmt.Sprintf("%.1fs ELAPSED  •  %.1f RPS", m.elapsed.Seconds(), rps)
	}
	state := mutedStyle.Render(rightText)
	if m.running {
		state = cyanStyle.Render(rightText)
	}

	dot := statusDotStyle.Render("●")
	if m.running && m.dotGlow {
		dot = statusDotGlowStyle.Render("●")
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center,
		dot, " ",
		titleStyle.Render("Pulse"),
		" ",
		mutedStyle.Render("terminal"),
	)

	line := lipgloss.JoinHorizontal(lipgloss.Center,
		left,
		lipgloss.NewStyle().Width(width-lipgloss.Width(left)-lipgloss.Width(state)).Render(""),
		state,
	)

	separator := separatorBorder.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, line, separator)
}

func (m Model) renderCommand(width int) string {
	methodName := runconfig.AllowedMethods()[m.methodIndex]
	method := methodStyle(methodName, m.focus == focusMethod).Render(methodName)
	methodBox := lipgloss.NewStyle().Width(10).Render(method)

	urlBox := inputStyle(m.focus == focusURL).
		Width(max(20, width-56)).
		Render(truncate(m.urlInput.Value(), max(18, width-58)))

	ccBox := controlStyle(m.focus == focusConcurrency).
		Width(8).
		Render(fmt.Sprintf("CC %d", m.concurrency()))

	payloadLabel := "PAYLOAD"
	if m.showPayload {
		payloadLabel = "PAYLOAD·ON"
	}
	payloadBox := controlStyle(m.focus == focusPayload).
		Width(11).
		Render(payloadLabel)

	runLabel := "▶ RUN"
	if m.running {
		runLabel = "◼ CANCEL"
	}
	runBox := runButtonStyle(m.running).Width(10).Render(runLabel)

	divider := separatorBorder.Render("│")

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		methodBox, " ", divider, " ",
		urlBox, " ", divider, " ",
		ccBox, " ", divider, " ",
		payloadBox, " ", divider, " ",
		runBox,
	)

	return panelStyle.Width(width).Render(row)
}

func (m Model) renderMetrics(width int) string {
	summary := m.summary
	segWidth := max(14, (width-6)/4)

	target := m.concurrency()
	barWidth := max(4, segWidth-8)
	var bar string
	if target > 0 {
		filled := summary.Total * barWidth / target
		if filled > barWidth {
			filled = barWidth
		}
		bar = strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		bar += fmt.Sprintf(" %d/%d", summary.Total, target)
	} else {
		bar = fmt.Sprintf("%d", summary.Total)
	}
	requests := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("REQUESTS"),
		metricValueStyle.Render(bar),
	)

	successColor := colorAmber
	if summary.SuccessRate >= 100 {
		successColor = colorGreen
	}
	success := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("SUCCESS"),
		lipgloss.NewStyle().Foreground(lipgloss.Color(successColor)).Bold(true).Render(fmt.Sprintf("%d%%", summary.SuccessRate)),
	)

	errors := summary.Total - summary.Successes
	errColor := colorGreen
	if errors > 0 {
		errColor = colorRose
	}
	rps := 0.0
	if m.elapsed > 0 {
		rps = float64(summary.Total) / m.elapsed.Seconds()
	}
	errorsRPS := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("ERRORS"),
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.Color(errColor)).Bold(true).Render(fmt.Sprintf("%d", errors)),
			" ",
			mutedStyle.Render(fmt.Sprintf("%.1f/s", rps)),
		),
	)

	lat := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("LATENCY"),
		metricValueStyle.Render(fmt.Sprintf("%s/%s",
			formatDuration(summary.P50),
			formatDuration(summary.P99),
		)),
	)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(segWidth).Render(requests), "  ",
		lipgloss.NewStyle().Width(segWidth).Render(success), "  ",
		lipgloss.NewStyle().Width(segWidth).Render(errorsRPS), "  ",
		lipgloss.NewStyle().Width(segWidth).Render(lat),
	)

	separator := separatorBorder.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, row, separator)
}

func (m Model) renderPayload(width int) string {
	leftWidth := max(28, width/2-2)
	rightWidth := max(28, width-leftWidth-3)

	headers := []string{sectionTitleStyle.Render("HEADERS  ctrl+n add  ctrl+d remove")}
	if len(m.headers) == 0 {
		headers = append(headers, mutedStyle.Render("No headers configured."))
	} else {
		for i, header := range m.headers {
			prefix := "  "
			if m.focus == focusHeaders && i == m.selectedHead {
				prefix = "▸ "
			}
			key := truncate(header.Key.Value(), leftWidth/2-4)
			value := truncate(header.Value.Value(), leftWidth/2-4)
			if key == "" {
				key = mutedStyle.Render("Header")
			}
			if value == "" {
				value = mutedStyle.Render("Value")
			}
			line := fmt.Sprintf("%s%-*s %s", prefix, leftWidth/2-2, key, value)
			headers = append(headers, line)
		}
	}

	bodyPreview := m.bodyInput.View()
	if strings.TrimSpace(m.bodyInput.Value()) == "" && m.focus != focusBody {
		bodyPreview = mutedStyle.Render(`{"name":"pulse"}`)
	}

	left := panelStyle.Width(leftWidth).Render(strings.Join(headers, "\n"))
	right := panelStyle.Width(rightWidth).Render(sectionTitleStyle.Render("BODY") + "\n" + bodyStyle(m.focus == focusBody).Width(rightWidth-4).Render(bodyPreview))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m Model) renderWorkspace(width int, height int) string {
	contentWidth := width - 6
	tabs := m.renderTabs(contentWidth)
	tabsHeight := lipgloss.Height(tabs)
	bodyHeight := max(4, height-4-tabsHeight)

	resultsWidth := contentWidth
	inspector := ""
	if m.inspector && contentWidth >= 108 {
		resultsWidth = contentWidth*58/100 - 1
		inspector = m.renderInspector(contentWidth-resultsWidth-1, bodyHeight)
	}

	var results string
	if m.activeTab == tabTimeline {
		results = m.renderTimeline(resultsWidth, bodyHeight)
	} else {
		results = m.renderLogs(resultsWidth, bodyHeight)
	}

	body := results
	if inspector != "" {
		body = lipgloss.JoinHorizontal(lipgloss.Top, results, " ", inspector)
	} else if m.inspector {
		body = lipgloss.JoinVertical(lipgloss.Left, results, m.renderInspector(contentWidth, max(6, bodyHeight/2)))
	}

	return panelStyle.Width(width).Height(height).Render(
		lipgloss.JoinVertical(lipgloss.Left, tabs, body),
	)
}

func (m Model) renderTabs(width int) string {
	timeline := pillStyle(m.activeTab == tabTimeline).Render("Timeline")
	logs := pillStyle(m.activeTab == tabLogs).Render("Live Logs")
	indicator := ""
	if !m.autoScroll {
		indicator = mutedStyle.Render("  MANUAL")
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top, timeline, " ", logs, indicator),
	)
}

func (m Model) renderTimeline(width int, height int) string {
	if len(m.results) == 0 {
		msg := "Waiting for results..."
		if !m.running {
			if strings.TrimSpace(m.urlInput.Value()) == "" {
				msg = "Enter a URL to begin"
			} else {
				msg = "Ctrl+R to run"
			}
		}
		return emptyStyle.Width(width).Height(height).Render(msg)
	}

	maxStart := max(0, len(m.results)-height)
	start := max(0, min(m.selected-height/2, maxStart))

	lines := make([]string, 0, min(len(m.results)-start, height))
	for i := start; i < len(m.results) && len(lines) < height; i++ {
		result := m.results[i]
		sel := m.focus == focusResults && i == m.selected
		lines = append(lines, m.renderTimelineRow(i, result, m.summary.MaxLatency, width, sel))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderTimelineRow(index int, result model.Result, maxLatency time.Duration, width int, selected bool) string {
	status := resultStatus(result)
	latency := formatDuration(result.Latency)
	method := result.RequestMethod
	if method == "" {
		method = runconfig.AllowedMethods()[m.methodIndex]
	}

	barWidth := max(6, width-38)
	filled := 0
	if maxLatency > 0 {
		filled = int(float64(result.Latency) / float64(maxLatency) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
	}

	barColor := statusColor(result.Status)
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(
		strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled),
	)

	line := fmt.Sprintf("%s %3d %-4s %-12s %s %s",
		rowCursor(selected), index+1, method, status, bar, latency)

	if result.Status >= 400 || result.Status == 0 {
		return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
	}
	return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
}

func (m Model) renderLogs(width int, height int) string {
	if len(m.results) == 0 {
		msg := "No results yet..."
		if !m.running {
			if strings.TrimSpace(m.urlInput.Value()) == "" {
				msg = "Enter a URL to begin"
			} else {
				msg = "Ctrl+R to run"
			}
		}
		return emptyStyle.Width(width).Height(height).Render(msg)
	}

	maxStart := max(0, len(m.results)-height)
	start := max(0, min(m.selected-height/2, maxStart))

	lines := make([]string, 0, min(len(m.results)-start, height))
	for i := start; i < len(m.results) && len(lines) < height; i++ {
		result := m.results[i]
		sel := m.focus == focusResults && i == m.selected
		status := resultStatus(result)
		method := result.RequestMethod
		if method == "" {
			method = runconfig.AllowedMethods()[m.methodIndex]
		}
		reqURL := result.RequestURL
		if reqURL == "" {
			reqURL = m.urlInput.Value()
		}
		line := fmt.Sprintf("%s %3d %-4s %-10s %-8s %s",
			rowCursor(sel), i+1, method, status, formatDuration(result.Latency), truncate(reqURL, width-33))
		if result.Status >= 400 || result.Status == 0 {
			lines = append(lines, strings.TrimSpace(errorRowStyle(sel).Render(truncate(line, width))))
		} else {
			lines = append(lines, strings.TrimSpace(rowStyle(sel).Render(truncate(line, width))))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderInspector(width int, height int) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return panelStyle.Width(width).Height(height).Render(mutedStyle.Render("No result selected."))
	}

	result := m.results[m.selected]
	method := result.RequestMethod
	if method == "" {
		method = runconfig.AllowedMethods()[m.methodIndex]
	}
	reqURL := result.RequestURL
	if reqURL == "" {
		reqURL = m.urlInput.Value()
	}

	lines := []string{
		sectionTitleStyle.Render("INSPECTOR"),
		fmt.Sprintf("  %s %s", methodStyle(method, false).Render(method), truncate(reqURL, width-12)),
		"",
		fmt.Sprintf("Status:  %s", resultStatus(result)),
		fmt.Sprintf("Latency: %s", formatDuration(result.Latency)),
	}
	if result.Error != "" {
		lines = append(lines, errorTextStyle.Render("Error: "+result.Error))
	}
	lines = append(lines, "", sectionTitleStyle.Render("HEADERS"))

	if len(result.ResponseHeaders) == 0 {
		lines = append(lines, mutedStyle.Render("No headers captured."))
	} else {
		keys := make([]string, 0, len(result.ResponseHeaders))
		for key := range result.ResponseHeaders {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			lines = append(lines, truncate(fmt.Sprintf("%s: %s", key, result.ResponseHeaders[key]), width-4))
		}
	}

	lines = append(lines, "", sectionTitleStyle.Render("BODY"))
	body := result.ResponseBody
	if body == "" {
		body = mutedStyle.Render("No body captured.")
	}
	bodyLines := strings.Split(body, "\n")
	for i, bline := range bodyLines {
		if len(lines) >= height-2 {
			if i < len(bodyLines)-1 {
				lines = append(lines, mutedStyle.Render("... (truncated)"))
			}
			break
		}
		lines = append(lines, truncate(bline, width-4))
	}

	return panelStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) renderFooter(width int) string {
	left := "tab/⇧tab focus  ctrl+r run  ctrl+x cancel  ctrl+a scroll  [/] tabs  ↑↓ select  enter inspect  esc back  q quit"

	right := ""
	if !m.running && len(m.results) > 0 {
		s := metrics.Compute(m.results, m.elapsed)
		right = fmt.Sprintf("%d results  p99 %s  %d errors",
			s.Total, formatDuration(s.P99), s.Total-s.Successes)
	}

	var line string
	if right != "" {
		leftWidth := max(40, width-lipgloss.Width(mutedStyle.Render(right))-4)
		line = lipgloss.JoinHorizontal(lipgloss.Center,
			mutedStyle.Width(leftWidth).Render(truncate(left, leftWidth)),
			mutedStyle.Render(right),
		)
	} else {
		line = mutedStyle.Render(truncate(left, width))
	}

	return separatorBorder.Render(strings.Repeat("─", width)) + "\n" + line
}

func displayElapsed(m Model) time.Duration {
	if m.running || m.elapsed > 0 {
		return m.elapsed
	}
	return 0
}

func (m Model) renderSparkline(width int) string {
	if m.latencyLen == 0 {
		return ""
	}

	label := lipgloss.PlaceHorizontal(width, lipgloss.Right, labelStyle.Render("LATENCY SPARKLINE"))

	count := m.latencyLen
	if count > width {
		count = width
	}

	maxLat := time.Duration(0)
	for i := 0; i < count; i++ {
		idx := (m.latencyHead - count + i + latencyRingSize) % latencyRingSize
		if m.latencyRing[idx] > maxLat {
			maxLat = m.latencyRing[idx]
		}
	}

	var sb strings.Builder
	for i := 0; i < count; i++ {
		idx := (m.latencyHead - count + i + latencyRingSize) % latencyRingSize
		level := 0
		if maxLat > 0 {
			level = int(float64(m.latencyRing[idx]) / float64(maxLat) * 7)
			if level < 0 {
				level = 0
			} else if level > 7 {
				level = 7
			}
		}
		dotColor := colorGreen
		if level >= 3 {
			dotColor = colorAmber
		}
		if level >= 6 {
			dotColor = colorRose
		}
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(dotColor)).Render(string(sparklineChars[level])))
	}

	bars := lipgloss.NewStyle().Width(width).MaxWidth(width).Render(sb.String())
	separator := separatorBorder.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, label, bars, separator)
}

func resultStatus(result model.Result) string {
	if result.Status == 0 {
		return "ERR"
	}
	switch {
	case result.Status >= 100 && result.Status < 200:
		return fmt.Sprintf("%d Info", result.Status)
	case result.Status >= 200 && result.Status < 300:
		return fmt.Sprintf("%d OK", result.Status)
	case result.Status >= 300 && result.Status < 400:
		return fmt.Sprintf("%d Redirect", result.Status)
	case result.Status >= 400:
		return fmt.Sprintf("%d", result.Status)
	default:
		return fmt.Sprintf("%d", result.Status)
	}
}

func rowCursor(selected bool) string {
	if selected {
		return "▸"
	}
	return " "
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = strings.ReplaceAll(value, "\n", " ")
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

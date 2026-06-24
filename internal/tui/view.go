package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Pulse is starting..."
	}

	width := max(72, m.width)
	bodyHeight := max(1, m.height-5)

	var sb strings.Builder
	sb.WriteString(m.renderTopBar(width))
	sb.WriteString("\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")

	switch {
	case m.dialog == dialogPayload:
		sb.WriteString(m.renderPayload(width))
	case m.dialog == dialogEndpoint:
		sb.WriteString(m.renderEndpoint(width))
	case m.dialog == dialogConcurrency:
		sb.WriteString(m.renderConcurrency(width))
	case m.mode == modeInspect:
		sb.WriteString(m.renderInspect(width, bodyHeight))
	case !m.running && len(m.results) == 0:
		sb.WriteString(m.renderReady(width, bodyHeight))
	case m.view == viewTimeline:
		sb.WriteString(m.renderTimeline(width, bodyHeight))
	default:
		sb.WriteString(m.renderLogs(width, bodyHeight))
	}

	sb.WriteString("\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(m.renderStatusBar(width))

	return styleBase.Width(width).Height(m.height).Render(sb.String())
}

func identityCell(label string, subdued bool) string {
	if subdued {
		return styleMuted.Render(label)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true).Render(label)
}

func (m Model) metricsString() string {
	if !m.running && len(m.results) == 0 {
		return ""
	}
	s := m.summary
	rps := 0.0
	if m.elapsed > 0 {
		rps = float64(s.Total) / m.elapsed.Seconds()
	}
	return fmt.Sprintf("%d%% ok • %.1f r/s • p90 %s • p99 %s",
		s.SuccessRate, rps, formatDuration(s.P90), formatDuration(s.P99))
}

func (m Model) renderTopBar(width int) string {
	method := runconfig.AllowedMethods()[m.methodIndex]
	url := truncateURL(m.urlInput.Value(), 40)
	left := method + " " + url

	cc := strings.TrimSpace(m.ccInput.Value())
	right := "CC " + cc

	maxLeft := width - lipgloss.Width(right) - 3
	if maxLeft < 12 {
		maxLeft = 12
	}
	leftTruncated := truncate(left, maxLeft)
	padding := width - lipgloss.Width(leftTruncated) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}

	line := leftTruncated + strings.Repeat(" ", padding) + right
	return styleTopBar.Width(width).Render(line)
}

func (m Model) renderReady(width int, height int) string {
	method := runconfig.AllowedMethods()[m.methodIndex]
	url := m.urlInput.Value()
	cc := m.concurrency()

	payloadParts := []string{}
	if len(m.headers) > 0 {
		payloadParts = append(payloadParts, fmt.Sprintf("%d header(s)", len(m.headers)))
	}
	if m.bodyInput.Value() != "" {
		payloadParts = append(payloadParts, "body present")
	}
	payloadState := "Empty"
	if len(payloadParts) > 0 {
		payloadState = strings.Join(payloadParts, ", ")
	}

	identity := identityCell("OBSERVE", false)

	content := fmt.Sprintf("%s    %s\n\nCC %d\n\nBody  %s",
		method, url, cc, payloadState)

	var b strings.Builder
	b.WriteString(identity)
	b.WriteString("\n\n")
	b.WriteString(content)

	return styleBase.Copy().Width(width).Height(height).Render(b.String())
}

func (m Model) renderEndpoint(width int) string {
	var b strings.Builder
	b.WriteString(identityCell("Endpoint", true))
	b.WriteString("\n")

	methods := runconfig.AllowedMethods()
	var methodLine string
	for i, method := range methods {
		sel := i == m.methodIndex
		focus := m.endpointField == endpointMethod
		if sel && focus {
			methodLine += lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Background(lipgloss.Color(colorDark)).Render(" " + method + " ")
		} else if sel {
			methodLine += lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render(" " + method + " ")
		} else {
			methodLine += styleMuted.Render(" " + method + " ")
		}
	}

	methodLabel := "Method"
	if m.endpointField == endpointMethod {
		methodLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render("Method")
	} else {
		methodLabel = styleMuted.Render("Method")
	}

	b.WriteString(fmt.Sprintf(
		"\n%s\n  %s\n%s\n  %s",
		methodLabel,
		methodLine,
		styleMuted.Render("URL"),
		m.urlInput.View(),
	))

	return b.String()
}

func (m Model) renderConcurrency(width int) string {
	var b strings.Builder
	b.WriteString(identityCell("Concurrency", true))
	b.WriteString("\n\n")

	ccText := strings.TrimSpace(m.ccInput.View())
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true).Render(
		fmt.Sprintf("  %s  (1–%d)", ccText, runconfig.MaxConcurrency),
	))
	return b.String()
}

func (m Model) renderPayload(width int) string {
	var b strings.Builder
	b.WriteString(identityCell("Payload", true))
	b.WriteString("\n")

	headersColor := colorMuted
	if m.selectedHead != bodyFocus {
		headersColor = colorAccent
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(headersColor)).Render("HEADERS  ctrl+n add  ctrl+d remove"))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(styleMuted.Render("  No headers configured."))
	} else {
		for i, header := range m.headers {
			key := header.Key.View()
			value := header.Value.View()
			sel := i == m.selectedHead
			cursor := rowCursor(sel)
			line := fmt.Sprintf("%s %s: %s", cursor, key, value)
			b.WriteString(rowStyle(sel).Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	bodyColor := colorMuted
	if m.selectedHead == bodyFocus {
		bodyColor = colorAccent
	}
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := "BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf("BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(bodyColor)).Render(bodyLabel))
	b.WriteString("\n")

	b.WriteString(m.bodyInput.View())

	return b.String()
}

func (m Model) renderStatusBar(width int) string {
	var hints string

	switch {
	case m.dialog == dialogConfirmQuit:
		hints = "Enter/q/Ctrl+C to quit, any key cancels"
	case m.dialog == dialogEndpoint:
		hints = "Tab Next Field  ← → Method  Esc Back"
	case m.dialog == dialogConcurrency:
		hints = "↑ ↓ Adjust  Esc Back"
	case m.dialog == dialogPayload:
		hints = "Tab Next  Ctrl+N Header  Ctrl+D Delete  Esc Back"
	case m.mode == modeInspect:
		hints = "↑ ↓ Select  Esc Back  q Quit"
	case m.running && len(m.results) == 0:
		hints = "Ctrl+X Cancel"
	case m.running:
		hints = "↑ ↓ Select  Enter Inspect  [ ] Views  Ctrl+X Cancel"
	case !m.running && len(m.results) == 0:
		hints = "e Endpoint  c Concurrency  p Payload  Ctrl+R Run  q Quit"
	default:
		hints = "↑ ↓ Select  Enter Inspect  [ ] Views  e Endpoint  c Concurrency  p Payload  Ctrl+R Run  q Quit"
	}

	return styleStatusBar.Width(width).Render(hints)
}

func visibleWindow(total, selected, height int) int {
	if total <= height {
		return 0
	}
	maxStart := total - height
	start := selected - height/2
	if start < 0 {
		start = 0
	}
	if start > maxStart {
		start = maxStart
	}
	return start
}

func (m Model) renderResultList(width, height int, identity string, emptyRunning string, rowFn func(result model.Result, index int, selected bool, width int) string) string {
	var b strings.Builder

	b.WriteString(identityCell(identity, false))
	b.WriteString("\n")

	remaining := height - 1
	if metrics := m.metricsString(); metrics != "" {
		b.WriteString(styleMuted.Width(width).Render(metrics))
		b.WriteString("\n")
		remaining--
	}

	if remaining <= 0 {
		return styleBase.Copy().Width(width).Height(height).Render(b.String())
	}

	if len(m.results) == 0 {
		msg := m.renderEmptyState(emptyRunning)
		b.WriteString(styleMuted.Render(msg))
		return styleBase.Copy().Width(width).Height(height).Render(b.String())
	}

	start := visibleWindow(len(m.results), m.selected, remaining)
	rows := make([]string, 0, min(len(m.results)-start, remaining))
	for i := start; i < len(m.results) && len(rows) < remaining; i++ {
		result := m.results[i]
		sel := i == m.selected
		rows = append(rows, rowFn(result, i, sel, width))
	}
	b.WriteString(strings.Join(rows, "\n"))

	return styleBase.Copy().Width(width).Height(height).Render(b.String())
}

func (m Model) renderTimeline(width int, height int) string {
	return m.renderResultList(width, height, "Timeline", "⏳  Waiting for results...",
		func(result model.Result, index int, selected bool, width int) string {
			return m.renderTimelineRow(index, result, m.summary.MaxLatency, width, selected)
		})
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
	return m.renderResultList(width, height, "Logs", "📭  No results yet...",
		func(result model.Result, index int, selected bool, width int) string {
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
				rowCursor(selected), index+1, method, status, formatDuration(result.Latency), truncate(reqURL, width-33))
			if result.Status >= 400 || result.Status == 0 {
				return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
			}
			return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
		})
}

func (m Model) renderInspect(width int, height int) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return styleBase.Copy().Width(width).Height(height).Render(styleMuted.Render("No result selected."))
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

	errorText := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
	sectionLine := func(label string) string {
		return styleMuted.Render("-- " + label + " --")
	}

	identity := identityCell(fmt.Sprintf("Inspector - Result #%d", m.selected+1), false)

	lines := []string{
		identity,
		"",
		fmt.Sprintf("  %s %s", method, truncate(reqURL, width-12)),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor(result.Status))).Render("Status:  " + resultStatus(result)),
		fmt.Sprintf("Latency: %s", formatDuration(result.Latency)),
	}
	if result.Error != "" {
		lines = append(lines, errorText.Render("Error: "+result.Error))
	}
	lines = append(lines, "", sectionLine("HEADERS"))

	if len(result.ResponseHeaders) == 0 {
		lines = append(lines, styleMuted.Render("No headers captured."))
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

	lines = append(lines, "", sectionLine("BODY"))
	body := result.ResponseBody
	if body == "" {
		body = styleMuted.Render("No body captured.")
	}
	bodyLines := strings.Split(body, "\n")
	for i, bline := range bodyLines {
		if len(lines) >= height-2 {
			if i < len(bodyLines)-1 {
				lines = append(lines, styleMuted.Render("... (truncated)"))
			}
			break
		}
		lines = append(lines, truncate(bline, width-4))
	}

	return styleBase.Copy().Width(width).Height(height).Render(strings.Join(lines, "\n"))
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
		return "▶"
	}
	return " "
}

func (m Model) renderEmptyState(runningMsg string) string {
	if !m.running {
		if strings.TrimSpace(m.urlInput.Value()) == "" {
			return "Enter a URL to begin"
		}
		return "▶  Ctrl+R to run"
	}
	return runningMsg
}

func truncateURL(rawURL string, width int) string {
	if width <= 0 || rawURL == "" {
		return rawURL
	}
	u := strings.TrimPrefix(rawURL, "https://")
	u = strings.TrimPrefix(u, "http://")
	if idx := strings.Index(u, "?"); idx >= 0 {
		u = u[:idx]
	}
	if len([]rune(u)) <= width {
		return u
	}
	return truncate(u, width)
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

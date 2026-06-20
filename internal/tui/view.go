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

	fixed := 5
	var payloadHeight int
	if m.showPayload {
		payloadHeight = lipgloss.Height(m.renderPayload(width))
		fixed += payloadHeight
	}

	metricsStr := m.renderMetrics(width)
	if metricsStr != "" {
		fixed++
	}
	workspaceHeight := max(1, m.height-fixed)

	var sb strings.Builder
	sb.WriteString(m.renderTopBar(width))
	sb.WriteString("\n")
	sb.WriteString(StyleSeparator.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	if metricsStr != "" {
		sb.WriteString(metricsStr)
		sb.WriteString("\n")
	}
	sb.WriteString(m.renderTabStrip(width))
	sb.WriteString("\n")
	if m.showPayload {
		sb.WriteString(m.renderPayload(width))
		sb.WriteString("\n")
	}

	switch m.activeTab {
	case tabTimeline:
		if m.inspector && width >= 86 {
			resultsWidth := width * 58 / 100
			inspWidth := width - resultsWidth - 1
			sb.WriteString(m.renderTimeline(resultsWidth, workspaceHeight))
			sb.WriteString(" ")
			sb.WriteString(m.renderInspector(inspWidth, workspaceHeight))
		} else {
			sb.WriteString(m.renderTimeline(width, workspaceHeight))
			if m.inspector {
				sb.WriteString("\n")
				sb.WriteString(m.renderInspector(width, max(6, workspaceHeight/2)))
			}
		}
	default:
		if m.inspector && width >= 86 {
			resultsWidth := width * 58 / 100
			inspWidth := width - resultsWidth - 1
			sb.WriteString(m.renderLogs(resultsWidth, workspaceHeight))
			sb.WriteString(" ")
			sb.WriteString(m.renderInspector(inspWidth, workspaceHeight))
		} else {
			sb.WriteString(m.renderLogs(width, workspaceHeight))
			if m.inspector {
				sb.WriteString("\n")
				sb.WriteString(m.renderInspector(width, max(6, workspaceHeight/2)))
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString(StyleSeparator.Render(strings.Repeat("─", width)))
	sb.WriteString("\n")
	sb.WriteString(m.renderStatusBar(width))

	return StyleBase.Width(width).Height(m.height).Render(sb.String())
}

func (m Model) renderTopBar(width int) string {
	method := runconfig.AllowedMethods()[m.methodIndex]
	url := truncateURL(m.urlInput.Value(), 40)
	if m.focus == focusURL {
		url = "[" + url + "]"
	}
	left := method + " " + url

	var state string
	switch {
	case m.confirmQuit:
		state = "● QUIT?"
	case m.inspector:
		state = "● INSPECTING"
	case m.running:
		state = fmt.Sprintf("%d req    ● RUNNING %.1fs", len(m.results), m.elapsed.Seconds())
	case m.status == "COMPLETE":
		state = fmt.Sprintf("%d req    ● COMPLETED %.1fs", len(m.results), m.elapsed.Seconds())
	case m.status != "" && m.status != "SYSTEM READY":
		cc := strings.TrimSpace(m.ccInput.Value())
		if m.focus == focusConcurrency {
			cc = "[" + cc + "]"
		}
		state = fmt.Sprintf("CC %s    %s", cc, m.renderTopBarStatus())
	default:
		cc := strings.TrimSpace(m.ccInput.Value())
		if m.focus == focusConcurrency {
			cc = "[" + cc + "]"
		}
		state = fmt.Sprintf("CC %s    ● IDLE", cc)
	}

	rightStyled := renderRightStyled(m.running, m.status, m.confirmQuit, state)

	maxLeft := width - lipgloss.Width(rightStyled) - 3
	if maxLeft < 16 {
		maxLeft = 16
	}
	leftTruncated := truncate(left, maxLeft)
	padding := width - lipgloss.Width(leftTruncated) - lipgloss.Width(rightStyled)
	if padding < 1 {
		padding = 1
	}

	line := leftTruncated + strings.Repeat(" ", padding) + rightStyled
	return StyleTopBar.Width(width).Render(line)
}

func (m Model) renderTopBarStatus() string {
	return "● " + m.status
}

func renderRightStyled(running bool, status string, confirmQuit bool, right string) string {
	if running || confirmQuit || status != "" && status != "SYSTEM READY" {
		return StyleBase.Copy().Foreground(lipgloss.Color(colorAccent)).Render(right)
	}
	return StyleBase.Copy().Foreground(lipgloss.Color(colorMuted)).Render(right)
}

func (m Model) renderMetrics(width int) string {
	if !m.running && len(m.results) == 0 {
		return ""
	}
	s := m.summary
	sep := StyleBase.Copy().Foreground(lipgloss.Color(colorMuted)).Render(" • ")

	rps := 0.0
	if m.elapsed > 0 {
		rps = float64(s.Total) / m.elapsed.Seconds()
	}

	okStr := fmt.Sprintf("%d%% ok", s.SuccessRate)
	rpsStr := fmt.Sprintf("%.1f r/s", rps)
	p90Str := fmt.Sprintf("p90 %s", formatDuration(s.P90))
	p99Str := fmt.Sprintf("p99 %s", formatDuration(s.P99))

	return StyleBase.Copy().Width(width).Render(okStr + sep + rpsStr + sep + p90Str + sep + p99Str)
}

func (m Model) renderTabStrip(width int) string {
	var timeline, logs string
	if m.activeTab == tabTimeline {
		timeline = StyleBase.Copy().Foreground(lipgloss.Color(colorAccent)).Render("▶ Timeline")
		logs = StyleBase.Copy().Foreground(lipgloss.Color(colorMuted)).Render("  Logs")
	} else {
		timeline = StyleBase.Copy().Foreground(lipgloss.Color(colorMuted)).Render("  Timeline")
		logs = StyleBase.Copy().Foreground(lipgloss.Color(colorAccent)).Render("▶ Logs")
	}
	sep := StyleBase.Copy().Foreground(lipgloss.Color(colorMuted)).Render(" • ")
	return StyleBase.Copy().Width(width).Render(timeline + " " + sep + " " + logs)
}

func (m Model) renderPayload(width int) string {
	var b strings.Builder

	headersColor := colorMuted
	if m.focus == focusHeaders {
		headersColor = colorAccent
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(headersColor)).Render("HEADERS  ctrl+n add  ctrl+d remove"))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render("  No headers configured."))
	} else {
		for _, header := range m.headers {
			key := header.Key.Value()
			value := header.Value.Value()
			if key == "" {
				key = "Header"
			}
			if value == "" {
				value = "Value"
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	b.WriteString("\n")

	bodyColor := colorMuted
	if m.focus == focusBody {
		bodyColor = colorAccent
	}
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := "BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf("BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(bodyColor)).Render(bodyLabel))
	b.WriteString("\n")

	bodyPreview := m.bodyInput.Value()
	if strings.TrimSpace(bodyPreview) == "" {
		bodyPreview = `{"name":"pulse"}`
	}
	b.WriteString(bodyPreview)

	return b.String()
}

func (m Model) renderStatusBar(width int) string {
	var mode, hints string

	switch {
	case m.confirmQuit:
		mode = ""
		hints = "PRESS Q AGAIN TO QUIT • any other key cancels"
	case m.inspector:
		mode = "INSPECTING"
		hints = "• ↑↓ scroll • ESC back • Q quit"
	case m.running:
		mode = "RUNNING"
		hints = "• TAB focus • ENTER inspect • ^X cancel • Q quit"
	case m.focus == focusURL:
		mode = "EDIT URL"
		hints = "• TAB next • ESC cancel"
	case m.focus == focusConcurrency:
		mode = "EDIT CC"
		hints = "• TAB next • ESC cancel"
	default:
		mode = "NORMAL"
		hints = "• TAB focus • ENTER inspect • ^R run • ^X cancel • Q quit"
	}

	if mode == "" {
		return StyleStatusBar.Width(width).Render(hints)
	}

	modeStyled := StyleStatusMode.Render(" " + mode + " ")
	hintsStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(hints)
	line := modeStyled + " " + hintsStyled
	return StyleStatusBar.Width(width).Render(line)
}

func (m Model) renderTimeline(width int, height int) string {
	if len(m.results) == 0 {
		msg := m.renderEmptyState("⏳  Waiting for results...")
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Width(width).Height(height).Render(msg)
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
		msg := m.renderEmptyState("📭  No results yet...")
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Width(width).Height(height).Render(msg)
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
		return StyleBase.Copy().Width(width).Height(height).Render(lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render("No result selected."))
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

	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorAccent))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	errorText := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))

	lines := []string{
		sectionHeader.Render("INSPECTOR"),
		fmt.Sprintf("  %s %s", method, truncate(reqURL, width-12)),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor(result.Status))).Render("Status:  " + resultStatus(result)),
		fmt.Sprintf("Latency: %s", formatDuration(result.Latency)),
	}
	if result.Error != "" {
		lines = append(lines, errorText.Render("Error: "+result.Error))
	}
	lines = append(lines, "", sectionHeader.Render("HEADERS"))

	if len(result.ResponseHeaders) == 0 {
		lines = append(lines, muted.Render("No headers captured."))
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

	lines = append(lines, "", sectionHeader.Render("BODY"))
	body := result.ResponseBody
	if body == "" {
		body = muted.Render("No body captured.")
	}
	bodyLines := strings.Split(body, "\n")
	for i, bline := range bodyLines {
		if len(lines) >= height-2 {
			if i < len(bodyLines)-1 {
				lines = append(lines, muted.Render("... (truncated)"))
			}
			break
		}
		lines = append(lines, truncate(bline, width-4))
	}

	return StyleBase.Copy().Width(width).Height(height).Render(strings.Join(lines, "\n"))
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

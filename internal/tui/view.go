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

func (m Model) View() string {
	if m.width == 0 {
		return "Pulse is starting..."
	}

	width := clamp(m.width, 72, 140)
	innerWidth := width - 4

	header := m.renderHeader(innerWidth)
	command := m.renderCommand(innerWidth)
	metricStrip := m.renderMetrics(innerWidth)
	payload := ""
	if m.showPayload {
		payload = m.renderPayload(innerWidth)
	}

	used := lipgloss.Height(header) + lipgloss.Height(command) + lipgloss.Height(metricStrip) + lipgloss.Height(payload) + 5
	workspaceHeight := max(8, m.height-used)
	workspace := m.renderWorkspace(innerWidth, workspaceHeight)
	footer := footerStyle.Width(innerWidth).Render("tab focus  ctrl+r run  ctrl+x cancel  [/] tabs  enter inspect  q quit")

	parts := []string{header, command, metricStrip}
	if payload != "" {
		parts = append(parts, payload)
	}
	parts = append(parts, workspace, footer)

	return appStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (m Model) renderHeader(width int) string {
	state := mutedStyle.Render(m.status)
	if m.running {
		state = cyanStyle.Render(fmt.Sprintf("%.1fs ELAPSED", m.elapsed.Seconds()))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center,
		statusDotStyle.Render("●"),
		" ",
		titleStyle.Render("Pulse"),
		" ",
		mutedStyle.Render("terminal"),
	)
	right := monoStyle.Render(state)
	gap := strings.Repeat(" ", max(1, width-lipgloss.Width(left)-lipgloss.Width(right)))
	return lipgloss.JoinHorizontal(lipgloss.Center, left, gap, right)
}

func (m Model) renderCommand(width int) string {
	method := controlStyle(m.focus == focusMethod).Width(10).Render(runconfig.AllowedMethods[m.methodIndex])
	url := inputStyle(m.focus == focusURL).Width(max(20, width-54)).Render(truncate(m.urlInput.Value(), max(18, width-56)))
	cc := controlStyle(m.focus == focusConcurrency).Width(8).Render(fmt.Sprintf("CC %d", m.concurrency()))

	payloadLabel := "PAYLOAD OFF"
	if m.showPayload {
		payloadLabel = "PAYLOAD ON"
	}
	payload := controlStyle(m.focus == focusPayload).Width(13).Render(payloadLabel)

	runLabel := "RUN"
	if m.running {
		runLabel = "CANCEL"
	}
	run := runButtonStyle(m.running).Width(8).Render(runLabel)

	row := lipgloss.JoinHorizontal(lipgloss.Top, method, " ", url, " ", cc, " ", payload, " ", run)
	return panelStyle.Width(width).Render(row)
}

func (m Model) renderMetrics(width int) string {
	summary := metrics.Compute(m.results, displayElapsed(m))
	segmentWidth := max(12, (width-6)/4)
	requests := metricStyle.Width(segmentWidth).Render(fmt.Sprintf("REQUESTS\n%d / %d", summary.Total, m.concurrency()))
	success := metricStyle.Width(segmentWidth).Render(fmt.Sprintf("SUCCESS\n%d%%", summary.SuccessRate))
	latency := metricStyle.Width(segmentWidth).Render(fmt.Sprintf("AVG LATENCY\n%s", formatDuration(summary.Average)))
	rps := metricStyle.Width(segmentWidth).Render(fmt.Sprintf("RPS\n%.1f", summary.RequestsPerS))
	return lipgloss.JoinHorizontal(lipgloss.Top, requests, "  ", success, "  ", latency, "  ", rps)
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
				prefix = "> "
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
	tabs := m.renderTabs(width)
	bodyHeight := max(4, height-lipgloss.Height(tabs)-2)

	resultsWidth := width
	inspector := ""
	if m.inspector && width >= 108 {
		resultsWidth = width*58/100 - 1
		inspector = m.renderInspector(width-resultsWidth-1, bodyHeight)
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
		body = lipgloss.JoinVertical(lipgloss.Left, results, m.renderInspector(width, max(6, bodyHeight/2)))
	}

	return panelStyle.Width(width).Height(height).Render(lipgloss.JoinVertical(lipgloss.Left, tabs, body))
}

func (m Model) renderTabs(width int) string {
	timeline := tabStyle(m.activeTab == tabTimeline).Render("Timeline")
	logs := tabStyle(m.activeTab == tabLogs).Render("Live Logs")
	return lipgloss.PlaceHorizontal(width-2, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, timeline, logs))
}

func (m Model) renderTimeline(width int, height int) string {
	if len(m.results) == 0 {
		return emptyStyle.Width(width - 4).Height(height).Render("Awaiting execution...")
	}

	summary := metrics.Compute(m.results, displayElapsed(m))
	lines := make([]string, 0, min(len(m.results), height))
	start := max(0, len(m.results)-height)
	for i := start; i < len(m.results); i++ {
		result := m.results[i]
		selected := m.focus == focusResults && i == m.selected
		lines = append(lines, m.renderTimelineRow(i, result, summary.MaxLatency, width-4, selected))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderTimelineRow(index int, result model.Result, maxLatency time.Duration, width int, selected bool) string {
	status := resultStatus(result)
	latency := formatDuration(result.Latency)
	labelWidth := 18
	barWidth := max(8, width-labelWidth-len(latency)-6)
	filled := 1
	if maxLatency > 0 {
		filled = max(1, int(float64(result.Latency)/float64(maxLatency)*float64(barWidth)))
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("-", max(0, barWidth-filled))
	line := fmt.Sprintf("%s %3d %-10s [%s] %8s", rowCursor(selected), index+1, status, bar, latency)
	if result.Status >= 400 || result.Status == 0 {
		return errorRowStyle(selected).Render(truncate(line, width))
	}
	return rowStyle(selected).Render(truncate(line, width))
}

func (m Model) renderLogs(width int, height int) string {
	if len(m.results) == 0 {
		return emptyStyle.Width(width - 4).Height(height).Render("No logs yet.")
	}

	lines := make([]string, 0, min(len(m.results), height))
	start := max(0, len(m.results)-height)
	for i := start; i < len(m.results); i++ {
		result := m.results[i]
		selected := m.focus == focusResults && i == m.selected
		status := resultStatus(result)
		line := fmt.Sprintf("%s %3d %-7s %-8s %s", rowCursor(selected), i+1, status, formatDuration(result.Latency), truncate(m.urlInput.Value(), width-28))
		if result.Status >= 400 || result.Status == 0 {
			lines = append(lines, errorRowStyle(selected).Render(truncate(line, width-4)))
		} else {
			lines = append(lines, rowStyle(selected).Render(truncate(line, width-4)))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderInspector(width int, height int) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return panelStyle.Width(width).Height(height).Render(mutedStyle.Render("No result selected."))
	}

	result := m.results[m.selected]
	lines := []string{
		sectionTitleStyle.Render("INSPECTOR"),
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
	for _, line := range strings.Split(body, "\n") {
		lines = append(lines, truncate(line, width-4))
		if len(lines) >= height-2 {
			break
		}
	}

	return panelStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func displayElapsed(m Model) time.Duration {
	if m.running {
		return m.elapsed
	}
	return m.elapsed
}

func resultStatus(result model.Result) string {
	if result.Status == 0 {
		return "ERR"
	}
	if result.Status >= 200 && result.Status < 400 {
		return fmt.Sprintf("%d OK", result.Status)
	}
	return fmt.Sprintf("%d", result.Status)
}

func rowCursor(selected bool) string {
	if selected {
		return ">"
	}
	return " "
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = strings.ReplaceAll(value, "\n", " ")
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	return value[:width-1] + "..."
}

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

type Region struct {
	Width  int
	Height int
}

type ShellLayout struct {
	Context   Region
	Workspace Region
	Command   Region
}

func computeShellLayout(totalWidth, totalHeight int) ShellLayout {
	width := max(72, totalWidth)
	bodyHeight := max(1, totalHeight-5)
	return ShellLayout{
		Context:   Region{Width: width, Height: 1},
		Workspace: Region{Width: width, Height: bodyHeight},
		Command:   Region{Width: width, Height: 1},
	}
}

// ShellColumnWidth is the fixed width reserved for the orientation label
// in the operator ribbon. Actions always start at Column 19
// (ShellColumnWidth + 2-char gutter + 1 space).
const ShellColumnWidth = 16

// ActionCategory defines the four semantic groups an action can belong to.
// Categories are ordered by priority. Empty categories are omitted.
type ActionCategory int

const (
	NavigationCategory ActionCategory = iota
	ConfigurationCategory
	OperationCategory
	ApplicationCategory
)

// ActionID identifies an operator intent. The workspace exposes which actions
// currently exist; the ribbon layer maps them to presentation.
type ActionID int

const (
	ActionSelect ActionID = iota
	ActionInspect
	ActionSwitchView
	ActionConfigureRequest
	ActionRun
	ActionCancel
	ActionNextField
	ActionSwitchMethod
	ActionAdjustConcurrency
	ActionNextHeader
	ActionAddHeader
	ActionDeleteHeader
	ActionBack
	ActionQuit
	ActionConfirmQuit
	ActionCtrlCQuit
	ActionQQuit
	ActionDismissCancel
)

// Action is a behavioral intent — not a presentation object.
// The workspace produces actions; the ribbon derives presentation from them.
type Action struct {
	ID       ActionID
	Category ActionCategory
	Enabled  bool
}

// configItem represents a single configuration value in the Context region.
// The contract is Identity, Value, Validity — presentation is a shell concern.
type configItem struct {
	Identity string
	Value    string
	Valid    bool
}

// ShellState is an immutable snapshot of the shell-level state produced once
// per frame. Every shell renderer consumes this instead of reading model
// fields independently, eliminating duplicate state lookups.
type ShellState struct {
	Orientation   string
	Configuration []configItem
	Actions       []Action
}

// actionBinding maps an operator intent (ActionID) to its presentation in the
// operator ribbon. This is the single source of truth for shortcut-to-intent
// mapping. All [Ctrl+R], [Esc], [Tab], etc. live here.
type actionBinding struct {
	Key      string
	Label    string
	Category ActionCategory
}

var actionBindings = map[ActionID]actionBinding{
	ActionSelect:            {"↑↓", "Select", NavigationCategory},
	ActionInspect:           {"Enter", "Inspect", NavigationCategory},
	ActionSwitchView:        {"Tab", "Views", NavigationCategory},
	ActionConfigureRequest:  {"e", "Request", ConfigurationCategory},
	ActionRun:               {"Ctrl+R", "Run", OperationCategory},
	ActionCancel:            {"Ctrl+X", "Cancel", OperationCategory},
	ActionNextField:         {"Tab", "Next Field", ConfigurationCategory},
	ActionSwitchMethod:      {"←→", "Method", ConfigurationCategory},
	ActionAdjustConcurrency: {"↑↓", "Adjust", ConfigurationCategory},
	ActionNextHeader:        {"Tab", "Next", ConfigurationCategory},
	ActionAddHeader:         {"Ctrl+N", "Header", ConfigurationCategory},
	ActionDeleteHeader:      {"Ctrl+D", "Delete", ConfigurationCategory},
	ActionBack:              {"Esc", "Back", ApplicationCategory},
	ActionQuit:              {"q", "Quit", ApplicationCategory},
	ActionConfirmQuit:       {"Enter", "Quit", ApplicationCategory},
	ActionCtrlCQuit:         {"Ctrl+C", "Quit", ApplicationCategory},
	ActionQQuit:             {"q", "Quit", ApplicationCategory},
	ActionDismissCancel:     {"Any", "Cancel", ApplicationCategory},
}

func (m Model) orientationLabel() string {
	switch {
	case m.dialog == dialogConfirmQuit:
		return "QUIT"
	case m.dialog == dialogRequest:
		return "REQUEST"
	case m.mode == modeInspect:
		return "INSPECT"
	default:
		return "OBSERVE"
	}
}

func (m Model) Actions() []Action {
	switch {
	case m.dialog == dialogConfirmQuit:
		return []Action{
			{ActionConfirmQuit, ApplicationCategory, true},
			{ActionCtrlCQuit, ApplicationCategory, true},
			{ActionQQuit, ApplicationCategory, true},
			{ActionDismissCancel, ApplicationCategory, true},
		}
	case m.dialog == dialogRequest:
		return m.requestActions()
	case m.mode == modeInspect:
		return []Action{
			{ActionSelect, NavigationCategory, true},
			{ActionBack, ApplicationCategory, true},
			{ActionQuit, ApplicationCategory, true},
		}
	case m.running && len(m.results) == 0:
		return []Action{
			{ActionCancel, OperationCategory, true},
		}
	case m.running:
		return []Action{
			{ActionSelect, NavigationCategory, true},
			{ActionInspect, NavigationCategory, true},
			{ActionSwitchView, NavigationCategory, true},
			{ActionCancel, OperationCategory, true},
		}
	case !m.running && len(m.results) == 0:
		return []Action{
			{ActionConfigureRequest, ConfigurationCategory, true},
			{ActionRun, OperationCategory, true},
			{ActionQuit, ApplicationCategory, true},
		}
	default:
		return []Action{
			{ActionSelect, NavigationCategory, true},
			{ActionInspect, NavigationCategory, true},
			{ActionSwitchView, NavigationCategory, true},
			{ActionConfigureRequest, ConfigurationCategory, true},
			{ActionRun, OperationCategory, true},
			{ActionQuit, ApplicationCategory, true},
		}
	}
}

func (m Model) requestActions() []Action {
	domainActions := []Action{
		{ActionBack, ApplicationCategory, true},
		{ActionRun, OperationCategory, true},
	}
	switch m.activeDomain {
	case domainRequest:
		domainActions = append([]Action{
			{ActionNextField, ConfigurationCategory, true},
			{ActionSwitchMethod, ConfigurationCategory, true},
		}, domainActions...)
	case domainPayload:
		domainActions = append([]Action{
			{ActionNextField, ConfigurationCategory, true},
			{ActionAddHeader, ConfigurationCategory, true},
			{ActionDeleteHeader, ConfigurationCategory, true},
		}, domainActions...)
	case domainExec:
		domainActions = append([]Action{
			{ActionAdjustConcurrency, ConfigurationCategory, true},
		}, domainActions...)
	}
	return domainActions
}

func (m Model) Configuration() []configItem {
	ps := m.payloadSummary()
	items := []configItem{
		{"Method", runconfig.AllowedMethods()[m.methodIndex], true},
		{"URL", m.urlInput.Value(), m.urlInput.Value() != ""},
		{"CC", strings.TrimSpace(m.ccInput.Value()), true},
	}
	if ps != "—" {
		items = append(items, configItem{"Payload", ps, true})
	}
	return items
}

func (m Model) ShellState() ShellState {
	return ShellState{
		Orientation:   m.orientationLabel(),
		Configuration: m.Configuration(),
		Actions:       m.Actions(),
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Pulse is starting..."
	}

	layout := computeShellLayout(m.width, m.height)
	state := m.ShellState()

	var sb strings.Builder
	sb.WriteString(m.renderTopBar(state, layout.Context.Width))
	sb.WriteString("\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", layout.Context.Width)))
	sb.WriteString("\n")
	sb.WriteString(m.renderCurrentSurface(layout.Workspace))
	sb.WriteString("\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", layout.Command.Width)))
	sb.WriteString("\n")
	sb.WriteString(m.renderRibbon(state, layout.Command.Width))

	return styleBase.Width(layout.Context.Width).Height(m.height).Render(sb.String())
}

func (m Model) renderCurrentSurface(region Region) string {
	switch {
	case m.dialog == dialogRequest:
		return m.renderRequest(region)
	case m.mode == modeInspect:
		return m.renderInspect(region)
	case !m.running && len(m.results) == 0:
		return m.renderReady(region)
	case m.view == viewTimeline:
		return m.renderTimeline(region)
	default:
		return m.renderLogs(region)
	}
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

func (m Model) payloadSummary() string {
	h := len(m.headers)
	hasBody := m.bodyInput.Value() != ""
	switch {
	case h == 0 && !hasBody:
		return "—"
	case h > 0 && hasBody:
		return fmt.Sprintf("%dH+B", h)
	case h > 0:
		return fmt.Sprintf("%dH", h)
	default:
		return "B"
	}
}

func (m Model) renderTopBar(state ShellState, width int) string {
	cfg := state.Configuration
	left := ""
	right := ""
	for _, c := range cfg {
		switch c.Identity {
		case "Method":
			left = c.Value
		case "URL":
			left += " " + truncateURL(c.Value, 40)
		case "CC":
			right = "CC " + c.Value
		case "Payload":
			if width >= 100 {
				right += " · Payload " + c.Value
			}
		}
	}

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

func (m Model) renderReady(region Region) string {
	method := runconfig.AllowedMethods()[m.methodIndex]
	url := m.urlInput.Value()
	cc := m.concurrency()

	payloadLabel := "Payload " + m.payloadSummary()

	identity := identityCell("OBSERVE", false)

	content := fmt.Sprintf("%s    %s\n\nCC %d\n\n%s",
		method, url, cc, payloadLabel)

	var b strings.Builder
	b.WriteString(identity)
	b.WriteString("\n\n")
	b.WriteString(content)

	return styleBase.Copy().Width(region.Width).Height(region.Height).Render(b.String())
}

func isErrorResult(result model.Result) bool {
	return result.Status >= 400 || result.Status == 0
}

func accentOrMuted(name string, active bool) string {
	if active {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render(name)
	}
	return styleMuted.Render(name)
}

func (m Model) renderRequestDomain(b *strings.Builder) {
	b.WriteString("\n")
	b.WriteString(accentOrMuted("Request", m.activeDomain == domainRequest))
	b.WriteString("\n")

	methods := runconfig.AllowedMethods()
	var methodLine string
	for i, method := range methods {
		sel := i == m.methodIndex
		focus := m.activeDomain == domainRequest && m.requestField == reqFieldMethod
		if sel && focus {
			methodLine += lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Background(lipgloss.Color(colorDark)).Render(" " + method + " ")
		} else if sel {
			methodLine += lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render(" " + method + " ")
		} else {
			methodLine += styleMuted.Render(" " + method + " ")
		}
	}

	methodLabel := "Method"
	if m.activeDomain == domainRequest && m.requestField == reqFieldMethod {
		methodLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render("Method")
	} else {
		methodLabel = styleMuted.Render("Method")
	}

	b.WriteString(fmt.Sprintf("  %s\n    %s\n", methodLabel, methodLine))

	urlLabel := "URL"
	if m.activeDomain == domainRequest && m.requestField == reqFieldURL {
		urlLabel = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render("URL")
	} else {
		urlLabel = styleMuted.Render("URL")
	}
	b.WriteString(fmt.Sprintf("  %s\n    %s\n", urlLabel, m.urlInput.View()))
}

func (m Model) renderPayloadDomain(b *strings.Builder) {
	b.WriteString("\n")
	b.WriteString(accentOrMuted("Payload", m.activeDomain == domainPayload))
	b.WriteString("\n")

	headersActive := m.activeDomain == domainPayload && m.selectedHead != bodyFocus
	b.WriteString(accentOrMuted("  HEADERS  ctrl+n add  ctrl+d remove", headersActive))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(styleMuted.Render("    No headers configured."))
	} else {
		for i, header := range m.headers {
			key := header.Key.View()
			value := header.Value.View()
			sel := i == m.selectedHead
			cursor := rowCursor(sel)
			line := fmt.Sprintf("  %s %s: %s", cursor, key, value)
			b.WriteString(rowStyle(sel).Render(line))
			b.WriteString("\n")
		}
	}

	bodyActive := m.activeDomain == domainPayload && m.selectedHead == bodyFocus
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := "  BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf("  BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(accentOrMuted(bodyLabel, bodyActive))
	b.WriteString("\n")
	b.WriteString("  " + m.bodyInput.View())
}

func (m Model) renderExecDomain(b *strings.Builder) {
	b.WriteString("\n")
	b.WriteString(accentOrMuted("Execution", m.activeDomain == domainExec))
	b.WriteString("\n")

	ccText := strings.TrimSpace(m.ccInput.View())
	active := m.activeDomain == domainExec
	ccLabel := accentOrMuted("  Concurrency", active)
	if active {
		b.WriteString(fmt.Sprintf("%s: %s  (1–%d)\n", ccLabel, ccText, runconfig.MaxConcurrency))
	} else {
		ccVal := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).Render(ccText)
		b.WriteString(fmt.Sprintf("%s: %s  (1–%d)\n", ccLabel, ccVal, runconfig.MaxConcurrency))
	}
}

func (m Model) renderRequest(region Region) string {
	var b strings.Builder
	b.WriteString(identityCell("REQUEST", false))
	b.WriteString("\n")

	m.renderRequestDomain(&b)
	m.renderPayloadDomain(&b)
	m.renderExecDomain(&b)

	return styleBase.Copy().Width(region.Width).Height(region.Height).Render(b.String())
}

func (m Model) renderRibbon(state ShellState, width int) string {
	label := state.Orientation
	actions := state.Actions

	type cmd struct {
		key      string
		label    string
		category ActionCategory
	}

	groups := make([][]cmd, 4)
	for _, a := range actions {
		b := actionBindings[a.ID]
		c := cmd{key: b.Key, label: b.Label, category: b.Category}
		groups[b.Category] = append(groups[b.Category], c)
	}

	var actionParts []string
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		var parts []string
		for _, c := range group {
			parts = append(parts, fmt.Sprintf("[%s] %s", c.key, c.label))
		}
		actionParts = append(actionParts, strings.Join(parts, " · "))
	}

	actionColumn := strings.Join(actionParts, "    ")

	anchor := lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Render("│")
	var ribbonParts []string
	ribbonParts = append(ribbonParts, fmt.Sprintf("%s %-*s", anchor, ShellColumnWidth-2, label))
	if actionColumn != "" {
		ribbonParts = append(ribbonParts, actionColumn)
	}
	ribbonLine := strings.Join(ribbonParts, "  ")

	if w := lipgloss.Width(ribbonLine); w < width {
		ribbonLine += strings.Repeat(" ", width-w)
	}

	return styleRibbon.Width(width).Render(ribbonLine)
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

func (m Model) renderResultList(region Region, identity string, emptyRunning string, rowFn func(result model.Result, index int, selected bool, width int) string) string {
	var b strings.Builder

	b.WriteString(identityCell(identity, false))
	b.WriteString("\n")

	remaining := region.Height - 1
	if metrics := m.metricsString(); metrics != "" {
		b.WriteString(styleMuted.Width(region.Width).Render(metrics))
		b.WriteString("\n")
		remaining--
	}

	if remaining <= 0 {
		return styleBase.Copy().Width(region.Width).Height(region.Height).Render(b.String())
	}

	if len(m.results) == 0 {
		msg := m.renderEmptyState(emptyRunning)
		b.WriteString(styleMuted.Render(msg))
		return styleBase.Copy().Width(region.Width).Height(region.Height).Render(b.String())
	}

	start := visibleWindow(len(m.results), m.selected, remaining)
	rows := make([]string, 0, min(len(m.results)-start, remaining))
	for i := start; i < len(m.results) && len(rows) < remaining; i++ {
		result := m.results[i]
		sel := i == m.selected
		rows = append(rows, rowFn(result, i, sel, region.Width))
	}
	b.WriteString(strings.Join(rows, "\n"))

	return styleBase.Copy().Width(region.Width).Height(region.Height).Render(b.String())
}

func (m Model) renderTimeline(region Region) string {
	return m.renderResultList(region, "Timeline", "⏳  Waiting for results...",
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

	if isErrorResult(result) {
		return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
	}
	return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
}

func (m Model) renderLogs(region Region) string {
	return m.renderResultList(region, "Logs", "📭  No results yet...",
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
			if isErrorResult(result) {
				return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
			}
			return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
		})
}

func (m Model) renderInspect(region Region) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return styleBase.Copy().Width(region.Width).Height(region.Height).Render(styleMuted.Render("No result selected."))
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
		fmt.Sprintf("  %s %s", method, truncate(reqURL, region.Width-12)),
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
			lines = append(lines, truncate(fmt.Sprintf("%s: %s", key, result.ResponseHeaders[key]), region.Width-4))
		}
	}

	lines = append(lines, "", sectionLine("BODY"))
	body := result.ResponseBody
	if body == "" {
		body = styleMuted.Render("No body captured.")
	}
	bodyLines := strings.Split(body, "\n")
	for i, bline := range bodyLines {
		if len(lines) >= region.Height-2 {
			if i < len(bodyLines)-1 {
				lines = append(lines, styleMuted.Render("... (truncated)"))
			}
			break
		}
		lines = append(lines, truncate(bline, region.Width-4))
	}

	return styleBase.Copy().Width(region.Width).Height(region.Height).Render(strings.Join(lines, "\n"))
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
		return "▶  Ready"
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

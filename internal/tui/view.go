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

const sentinelEmpty = "—"

const (
	gapSection      = "\n\n"
	indentField     = "  "
	indentNested    = "    "
	inlineSeparator = " · "
)

func (m Model) Actions() []Action {
	switch {
	case m.workspace.dialog == dialogConfirmQuit:
		return []Action{
			{ActionConfirmQuit, ApplicationCategory, true},
			{ActionCtrlCQuit, ApplicationCategory, true},
			{ActionQQuit, ApplicationCategory, true},
			{ActionDismissCancel, ApplicationCategory, true},
		}
	case m.workspace.dialog == dialogRequest:
		return m.requestActions()
	case m.workspace.mode == modeInspect:
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
	if d, ok := domainRegistry[m.activeDomain]; ok {
		domainActions = append(d.Actions(m), domainActions...)
	}
	return domainActions
}

func (m Model) Configuration() []configItem {
	ps := m.payloadSummary()
	items := []configItem{
		{"Method", runconfig.AllowedMethods()[m.methodIndex], true},
		{"URL", m.urlInput.Value(), m.urlInput.Value() != ""},
		{"Concurrency", strings.TrimSpace(m.concurrencyInput.Value()), true},
	}
	if ps != sentinelEmpty {
		items = append(items, configItem{"Payload", ps, true})
	}
	return items
}

func (m Model) ShellState() ShellState {
	return ShellState{
		Orientation:   orientationLabel(m),
		Configuration: m.Configuration(),
		Actions:       m.Actions(),
	}
}

const contextThreshold = 140
const contextMinWidth = 28

func (m Model) View() string {
	w, h := m.shell.Dimensions()
	if w == 0 {
		return "Pulse is starting..."
	}

	layout := m.shell.Layout()
	state := m.ShellState()
	orientation := state.Orientation

	var sb strings.Builder
	sb.WriteString(m.renderTopBar(state, layout.Context.Width))
	sb.WriteString("\n")
	sb.WriteString(styleSeparator.Render(strings.Repeat("─", layout.Context.Width)))
	sb.WriteString("\n")

	ws := layout.Workspace
	ws.Border = BorderFull
	ws.Title = orientation
	ws.Padding = 1
	inner := Region{Type: WorkspaceRegion, Width: ws.Width - 2 - 2*ws.Padding, Height: ws.Height - 2}
	sb.WriteString(ws.RenderBordered(m.renderWorkspaceContent(inner, w)))
	sb.WriteString("\n")

	sb.WriteString(styleSeparator.Render(strings.Repeat("─", layout.Command.Width)))
	sb.WriteString("\n")
	sb.WriteString(m.renderStatusline(state, layout.Command.Width))

	return styleBase.Width(layout.Context.Width).Height(h).Render(sb.String())
}

func (m Model) renderWorkspaceContent(region Region, width int) string {
	context := m.renderContextRegion(region)
	if context == "" || width < contextThreshold {
		return m.resolveSurface().Render(region)
	}

	ctxWidth := min(contextMinWidth, max(contextMinWidth, region.Width/3))
	primaryWidth := region.Width - ctxWidth - 1
	if primaryWidth < 40 {
		return m.resolveSurface().Render(region)
	}

	primary := m.resolveSurface().Render(Region{Type: WorkspaceRegion, Width: primaryWidth, Height: region.Height})
	contextPanel := m.renderContextRegion(Region{Type: ContextRegion, Width: ctxWidth, Height: region.Height})

	return lipgloss.JoinHorizontal(lipgloss.Top, primary, " ", contextPanel)
}

func (m Model) renderContextRegion(region Region) string {
	switch {
	case m.workspace.dialog == dialogRequest:
		return m.renderRequestContext(region)
	case m.workspace.mode == modeInspect && len(m.results) > 0:
		return m.renderInspectContext(region)
	case len(m.results) > 0:
		return m.renderObserveContext(region)
	default:
		return ""
	}
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
	b.WriteString(fmt.Sprintf(indentField+"%s %s\n", method, truncateURL(reqURL, region.Width-12)))
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
	b.WriteString(fmt.Sprintf(indentField+"URL: %s\n", truncateURL(m.urlInput.Value(), region.Width-8)))
	b.WriteString(fmt.Sprintf(indentField+"C: %d\n", m.concurrency()))
	b.WriteString(fmt.Sprintf(indentField+"Payload: %s\n", m.payloadSummary()))
	return regionStyle(region).Render(b.String())
}

func identityCell(label string) string {
	return styleMuted.Render(label)
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
		return sentinelEmpty
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
		case "Concurrency":
			right = "C " + c.Value
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

	payloadLabel := m.payloadSummary()

	var b strings.Builder
	b.WriteString(styleMuted.Render("Ready"))
	b.WriteString(gapSection)
	b.WriteString(accentOrMuted("Current Request", true))
	b.WriteString(gapSection)
	b.WriteString("Method\n" + method + gapSection)
	b.WriteString("URL\n" + url + gapSection)
	b.WriteString(fmt.Sprintf("Concurrency\n%d"+gapSection, cc))
	b.WriteString("Payload\n" + payloadLabel + gapSection)
	b.WriteString("Status\n")
	b.WriteString(styleMuted.Render("Ready to execute"))

	return regionStyle(region).Render(b.String())
}

func isErrorResult(result model.Result) bool {
	return result.Status >= 400 || result.Status == 0
}

func accentOrMuted(name string, active bool) string {
	if active {
		return styleAccent.Render(name)
	}
	return styleMuted.Render(name)
}

func domainHeader(label string, width int, active bool) string {
	if width <= 0 {
		width = 80
	}
	label = " " + label + " "
	ruleLen := width - lipgloss.Width(label)
	if ruleLen < 4 {
		return strings.TrimSpace(label)
	}
	ruleChar := "─"
	if active {
		ruleChar = "━"
	}
	left := strings.Repeat(ruleChar, ruleLen/2)
	right := strings.Repeat(ruleChar, ruleLen-ruleLen/2)
	line := left + label + right
	if active {
		return styleDomainActive.Render(line)
	}
	return styleDomainInactive.Render(line)
}

func (m Model) renderRequest(region Region) string {
	var b strings.Builder
	b.WriteString(renderWorkspaceBadge("REQUEST"))
	b.WriteString("\n")
	b.WriteString(m.renderRequestDomain(region.Width))
	b.WriteString(m.renderPayloadDomain(region.Width))
	b.WriteString(m.renderExecDomain(region.Width))

	return regionStyle(region).Render(b.String())
}

func (m Model) renderRequestDomain(width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(domainHeader("Request", width, m.activeDomain == DomainRequest))
	b.WriteString("\n")

	methods := runconfig.AllowedMethods()
	var methodLine string
	for i, method := range methods {
		sel := i == m.methodIndex
		focus := m.activeDomain == DomainRequest && m.requestField == reqFieldMethod
		methodLine += renderMethodPill(method, sel, focus)
	}

	methodLabel := "Method"
	if m.activeDomain == DomainRequest && m.requestField == reqFieldMethod {
		methodLabel = styleAccent.Render("Method")
	} else {
		methodLabel = styleMuted.Render("Method")
	}

	b.WriteString(fmt.Sprintf(indentField+"%s\n"+indentNested+"%s\n", methodLabel, methodLine))

	urlLabel := "URL"
	if m.activeDomain == DomainRequest && m.requestField == reqFieldURL {
		urlLabel = styleAccent.Render("URL")
	} else {
		urlLabel = styleMuted.Render("URL")
	}
	b.WriteString(fmt.Sprintf(indentField+"%s\n"+indentNested+"%s\n", urlLabel, m.urlInput.View()))

	return b.String()
}

func (m Model) renderPayloadDomain(width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(domainHeader("Payload", width, m.activeDomain == DomainPayload))
	b.WriteString("\n")

	headersActive := m.activeDomain == DomainPayload && m.selectedHead != bodyFocus
	b.WriteString(accentOrMuted(indentField+"HEADERS  ctrl+n add  ctrl+d remove", headersActive))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(styleMuted.Render(indentNested + "No headers configured."))
	} else {
		for i, header := range m.headers {
			key := header.Key.View()
			value := header.Value.View()
			sel := i == m.selectedHead
			cursor := rowCursor(sel)
			line := fmt.Sprintf(indentField+"%s %s: %s", cursor, key, value)
			b.WriteString(rowStyle(sel).Render(line))
			b.WriteString("\n")
		}
	}

	bodyActive := m.activeDomain == DomainPayload && m.selectedHead == bodyFocus
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := indentField + "BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf(indentField+"BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(accentOrMuted(bodyLabel, bodyActive))
	b.WriteString("\n")
	b.WriteString(indentField + m.bodyInput.View())
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderExecDomain(width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(domainHeader("Execution", width, m.activeDomain == DomainExec))
	b.WriteString("\n")

	ccText := strings.TrimSpace(m.concurrencyInput.View())
	active := m.activeDomain == DomainExec
	ccLabel := accentOrMuted(indentField+"Concurrency", active)
	if active {
		b.WriteString(fmt.Sprintf("%s: %s  (1-%d)\n", ccLabel, ccText, runconfig.MaxConcurrency))
	} else {
		ccVal := styleBase.Foreground(lipgloss.Color(colorText)).Render(ccText)
		b.WriteString(fmt.Sprintf("%s: %s  (1-%d)\n", ccLabel, ccVal, runconfig.MaxConcurrency))
	}

	return b.String()
}

func renderWorkspaceBadge(label string) string {
	return styleWorkspaceBadge.Render(label)
}

func renderMethodPill(method string, selected bool, focused bool) string {
	switch {
	case selected && focused:
		return stylePrimaryAction.Render(" " + method + " ")
	case selected:
		return styleAccent.Render(" " + method + " ")
	default:
		return styleMuted.Render(" " + method + " ")
	}
}

func renderStatusBadge(result model.Result) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor(result.Status))).
		Bold(true).
		Render("Status: " + resultStatus(result))
}

func renderLatencyBar(filled, barWidth int, barColor string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(barColor)).
		Render(strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled))
}

func renderInteractionStatus(m Model) string {
	switch {
	case m.workspace.dialog == dialogRequest:
		return "Editing"
	case m.workspace.dialog == dialogConfirmQuit:
		return "Quitting"
	case m.workspace.mode == modeInspect:
		return "Inspecting"
	case m.running:
		return "Running"
	case len(m.results) > 0:
		return "Completed"
	default:
		return "Ready"
	}
}

// Density controls how much information the ribbon displays at a given width.
// Higher density values pack more information into less space until the
// action strip is fully hidden. Badge and status are never removed.
type Density int

const (
	DensityFull            Density = iota // 0: 4-space gaps, "  │  " seps, full labels
	DensityRelaxed                        // 1: 1-space group gaps, "  │  " seps
	DensityCompact                        // 2: 1-space gaps, "│" seps
	DensityCompressed                     // 3: 1-space gaps, "│" seps, " " within-group
	DensityAbbreviated                    // 4: same as 3, abbrev labels
	DensityPriorityReduced                // 5: same as 4, drop Low priority
	DensityMinimal                        // 6: keys only
	DensityEmergency                      // 7: no actions
)

// RibbonLayout holds the pre-rendered components of the footer statusline.
type RibbonLayout struct {
	Badge   string
	Actions string
	Status  string
	Density Density
}

func abbreviateLabel(label string) string {
	abbr := map[string]string{
		"Request":   "Req",
		"Execution": "Exec",
		"Observe":   "Obs",
		"Header":    "Hdr",
		"Headers":   "Hdrs",
		"Delete":    "Del",
		"Select":    "Sel",
		"Inspect":   "Insp",
		"Cancel":    "Ccl",
		"Adjust":    "Adj",
		"Method":    "Mth",
	}
	if v, ok := abbr[label]; ok {
		return v
	}
	if idx := strings.Index(label, " "); idx != -1 {
		return label[:idx]
	}
	return label
}

func buildActionStrip(actions []Action, level Density) string {
	if level >= DensityEmergency {
		return ""
	}

	groupGap := indentNested
	withinSep := inlineSeparator
	if level >= DensityRelaxed {
		groupGap = " "
	}
	if level >= DensityCompressed {
		withinSep = " "
	}

	groups := make([][]actionBinding, 4)
	for _, a := range actions {
		b := actionBindings[a.ID]
		if level == DensityPriorityReduced && b.Priority == PriorityLow {
			continue
		}
		groups[b.Category] = append(groups[b.Category], b)
	}

	highlighted := false
	var actionParts []string
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		var parts []string
		for _, b := range group {
			var part string
			switch {
			case level >= DensityMinimal:
				part = fmt.Sprintf("[%s]", b.Key)
			case level >= DensityAbbreviated:
				part = fmt.Sprintf("[%s] %s", b.Key, abbreviateLabel(b.Label))
			default:
				part = fmt.Sprintf("[%s] %s", b.Key, b.Label)
			}
			if !highlighted {
				part = stylePrimaryAction.Render(" " + part)
				highlighted = true
			}
			parts = append(parts, part)
		}
		actionParts = append(actionParts, strings.Join(parts, withinSep))
	}
	return strings.Join(actionParts, groupGap)
}

func chooseRibbonLevel(badge, status string, actions []Action, width int) (Density, string) {
	badgeWidth := lipgloss.Width(badge)
	statusWidth := lipgloss.Width(status)

	for level := DensityFull; level <= DensityEmergency; level++ {
		actionText := buildActionStrip(actions, level)
		actionWidth := lipgloss.Width(actionText)

		sepWidth := 11
		if level >= DensityCompact {
			sepWidth = 2
		}

		if badgeWidth+statusWidth+sepWidth+actionWidth <= width {
			return level, actionText
		}
	}
	return DensityEmergency, ""
}

func renderRibbon(layout RibbonLayout, width int) string {
	sep := styleMuted.Render("│")
	var leftSep, rightSep string
	if layout.Density >= DensityCompact {
		leftSep = sep
		rightSep = sep
	} else {
		leftSep = "  " + sep + "  "
		rightSep = "  " + sep + "  "
	}

	ls := lipgloss.Width(leftSep)
	rs := lipgloss.Width(rightSep)
	padding := max(0, width-lipgloss.Width(layout.Badge)-lipgloss.Width(layout.Status)-ls-rs-lipgloss.Width(layout.Actions))

	line := layout.Badge + leftSep + layout.Actions + strings.Repeat(" ", padding) + rightSep + layout.Status
	return styleRibbon.Width(width).Render(line)
}

func (m Model) renderStatusline(state ShellState, width int) string {
	badge := renderWorkspaceBadge(state.Orientation)
	status := styleStatusCell.Render(renderInteractionStatus(m))
	level, actions := chooseRibbonLevel(badge, status, state.Actions, width)

	return renderRibbon(RibbonLayout{Badge: badge, Actions: actions, Status: status, Density: level}, width)
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
	return m.renderResultList(region, "Timeline", "⏳  Waiting for results...",
		func(result model.Result, index int, selected bool, width int) string {
			return m.renderTimelineRow(index, result, m.summary.MaxLatency, width, selected)
		})
}

func (m Model) renderTimelineRow(index int, result model.Result, maxLatency time.Duration, width int, selected bool) string {
	status := resultStatus(result)
	latency := formatDuration(result.Latency)
	method := m.effectiveMethod(result)

	barWidth := max(6, width-34)
	filled := 0
	if maxLatency > 0 {
		filled = int(float64(result.Latency) / float64(maxLatency) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
	}

	barColor := statusColor(result.Status)
	bar := renderLatencyBar(filled, barWidth, barColor)

	line := fmt.Sprintf("%s %-4s %-12s %s %s",
		rowCursor(selected), method, status, bar, latency)

	if isErrorResult(result) {
		return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
	}
	return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
}

func (m Model) renderLogs(region Region) string {
	return m.renderResultList(region, "Logs", "📭  No results yet...",
		func(result model.Result, index int, selected bool, width int) string {
			status := resultStatus(result)
			method := m.effectiveMethod(result)
			reqURL := m.effectiveURL(result)
			line := fmt.Sprintf("%s %-4s %-10s %-8s %s",
				rowCursor(selected), method, status, formatDuration(result.Latency), truncate(reqURL, width-29))
			if isErrorResult(result) {
				return strings.TrimSpace(errorRowStyle(selected).Render(truncate(line, width)))
			}
			return strings.TrimSpace(rowStyle(selected).Render(truncate(line, width)))
		})
}

func (m Model) renderInspect(region Region) string {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		return regionStyle(region).Render(styleMuted.Render("No result selected."))
	}

	result := m.results[m.selected]
	method := m.effectiveMethod(result)
	reqURL := m.effectiveURL(result)

	sectionLine := func(label string) string {
		return styleSectionLine.Render("── " + label + " ──")
	}

	identity := identityCell(fmt.Sprintf("Result %d", m.selected+1))

	lines := []string{
		identity,
		gapSection,
		fmt.Sprintf("%s %s", method, truncate(reqURL, region.Width-4)),
		renderStatusBadge(result),
		fmt.Sprintf("Latency: %s", formatDuration(result.Latency)),
	}
	if result.Error != "" {
		lines = append(lines, styleError.Render("Error: "+result.Error))
	}
	lines = append(lines, gapSection, sectionLine("HEADERS"))

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

	lines = append(lines, gapSection, sectionLine("BODY"))
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

	return regionStyle(region).Render(strings.Join(lines, "\n"))
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

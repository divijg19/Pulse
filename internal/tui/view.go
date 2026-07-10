package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
)

const (
	timelineFixedWidth = 34
	logsFixedWidth     = 29
	logsFixedSuffix    = 7
	contextRowWidth    = 12
	contextURLWidth    = 8
	topBarURLWidth     = 40
	topBarMinLeft      = 12
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
	case m.workspace.mode == modeCompare:
		return []Action{
			{ActionClear, ApplicationCategory, true},
			{ActionSwap, NavigationCategory, true},
			{ActionZoneNext, NavigationCategory, true},
			{ActionZoneScroll, NavigationCategory, true},
			{ActionBack, ApplicationCategory, true},
			{ActionQuit, ApplicationCategory, true},
		}
	case m.workspace.mode == modeInspect:
		return []Action{
			{ActionZoneNext, NavigationCategory, true},
			{ActionZoneScroll, NavigationCategory, true},
			{ActionCompare, NavigationCategory, true},
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

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		out.WriteByte(s[i])
	}
	return out.String()
}

func (m Model) Configuration() []configItem {
	ps := m.payloadSummary()
	items := []configItem{
		{"Method", runconfig.AllowedMethods()[m.methodIndex], true},
		{"URL", m.urlInput.Value(), m.urlInput.Value() != ""},
		{"CC", strings.TrimSpace(m.concurrencyInput.Value()), true},
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

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(m.renderTopBar(state, layout.Context.Width))
	sb.WriteString("\n")

	ws := layout.Workspace
	ws.Border = BorderFull
	ws.Title = state.Orientation
	ws.PaddingX = 1
	ws.PaddingY = 1
	sb.WriteString(ws.RenderBordered(m.renderWorkspaceContent(ws.ContentRegion(), w)))
	sb.WriteString("\n")
	sb.WriteString(m.renderStatusline(state, layout.Command.Width))

	return styleBase.Width(layout.Context.Width).Height(h).Render(sb.String())
}

func (m Model) renderWorkspaceContent(region Region, width int) string {
	context := m.renderContextRegion(region)
	if context == "" || width < contextThreshold {
		return m.resolveSurface().Render(region)
	}

	ctxWidth := region.Width / 3
	if ctxWidth < contextMinWidth {
		ctxWidth = contextMinWidth
	}
	primaryWidth := region.Width - ctxWidth - 1
	if primaryWidth < 40 {
		return m.resolveSurface().Render(region)
	}

	primary := m.resolveSurface().Render(Region{Type: WorkspaceRegion, Width: primaryWidth, Height: region.Height})
	contextPanel := m.renderContextRegion(Region{Type: ContextRegion, Width: ctxWidth, Height: region.Height})

	return lipgloss.JoinHorizontal(lipgloss.Top, primary, " ", contextPanel)
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
	const leftPad = " "
	cfg := state.Configuration
	left := ""
	right := ""
	for _, c := range cfg {
		switch c.Identity {
		case "Method":
			left = c.Value
		case "URL":
			left += " " + truncateURL(c.Value, topBarURLWidth)
		case "CC":
			right = " CC " + c.Value
		case "Payload":
			if width >= 100 {
				right += " · Payload " + c.Value
			}
		}
	}

	maxLeft := width - 1 - lipgloss.Width(right) - 3
	if maxLeft < topBarMinLeft {
		maxLeft = 12
	}
	leftTruncated := truncate(left, maxLeft)
	padding := width - 1 - lipgloss.Width(leftTruncated) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}

	line := leftPad + leftTruncated + strings.Repeat(" ", padding) + right
	return styleTopBar.Width(width).Render(line)
}

func isErrorResult(result model.Result) bool {
	return result.Status == 0 || ClassifyStatus(result.Status) >= StatusClientError
}

func accentOrMuted(name string, active bool) string {
	if active {
		return styleAccent.Render(name)
	}
	return styleMuted.Render(name)
}

func renderWorkspaceBadge(label string) string {
	return styleWorkspaceBadge.Render(label)
}

func renderInteractionStatus(m Model) string {
	switch {
	case m.errMsg != "":
		return m.errMsg
	case m.workspace.dialog == dialogRequest:
		return "Editing"
	case m.workspace.dialog == dialogConfirmQuit:
		return "Quitting"
	case m.workspace.mode == modeCompare:
		return "Comparing"
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

		sepWidth := 10
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
	badgeWidth := lipgloss.Width(layout.Badge)
	actionWidth := lipgloss.Width(layout.Actions)
	statusWidth := lipgloss.Width(layout.Status)

	availForStatus := max(0, width-badgeWidth-ls-rs-actionWidth)
	if statusWidth > availForStatus && availForStatus >= 1 {
		raw := stripANSI(layout.Status)
		truncated := truncate(raw, availForStatus)
		layout.Status = styleStatusCell.Render(truncated)
		statusWidth = lipgloss.Width(layout.Status)
	}

	padding := max(0, width-badgeWidth-statusWidth-ls-rs-actionWidth)

	var rb strings.Builder
	rb.WriteString(layout.Badge)
	rb.WriteString(leftSep)
	rb.WriteString(layout.Actions)
	rb.WriteString(strings.Repeat(" ", padding))
	rb.WriteString(rightSep)
	rb.WriteString(layout.Status)
	return styleRibbon.Width(width).Render(rb.String())
}

func (m Model) renderStatusline(state ShellState, width int) string {
	badge := renderWorkspaceBadge(state.Orientation)
	statusText := renderInteractionStatus(m)
	status := styleStatusCell.Render(statusText)
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

func resultStatus(result model.Result) string {
	if result.Status == 0 {
		return "ERR"
	}
	switch ClassifyStatus(result.Status) {
	case StatusInfo:
		return fmt.Sprintf("%d Info", result.Status)
	case StatusSuccess:
		return fmt.Sprintf("%d OK", result.Status)
	case StatusRedirect:
		return fmt.Sprintf("%d Redirect", result.Status)
	case StatusClientError, StatusServerError:
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

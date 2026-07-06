package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/divijg19/Pulse/internal/runconfig"
)

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

	methodActive := m.activeDomain == DomainRequest && m.requestField == reqFieldMethod
	methodLabel := accentOrMuted("Method", methodActive)
	b.WriteString(fmt.Sprintf(indentField+"%s\n"+indentNested+"%s\n", methodLabel, methodLine))

	urlActive := m.activeDomain == DomainRequest && m.requestField == reqFieldURL
	urlLabel := accentOrMuted("URL", urlActive)
	b.WriteString(fmt.Sprintf(indentField+"%s\n"+indentNested+"%s\n", urlLabel, m.urlInput.View()))

	return b.String()
}

func (m Model) renderPayloadDomain(width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(domainHeader("Payload", width, m.activeDomain == DomainPayload))
	b.WriteString("\n")

	geo := calculatePayloadGeometry(width)

	headersActive := m.activeDomain == DomainPayload && m.selectedHead != bodyFocus
	b.WriteString(accentOrMuted(indentField+"HEADERS", headersActive))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(styleMuted.Render(indentNested + "No headers configured."))
		b.WriteString("\n")
	} else {
		for i := range m.headers {
			key := m.headers[i].Key.View()
			value := m.headers[i].Value.View()
			sel := i == m.selectedHead
			cursor := rowCursor(sel)
			line := fmt.Sprintf(indentNested+"%s %s: %s", cursor, key, value)
			b.WriteString(rowStyle(sel).Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	bodyActive := m.activeDomain == DomainPayload && m.selectedHead == bodyFocus
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := "BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf("BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(indentField + accentOrMuted(bodyLabel, bodyActive))
	b.WriteString("\n")
	m.bodyInput.SetWidth(geo.BodyWidth)
	bodyView := m.bodyInput.View()
	bodyLines := strings.Split(bodyView, "\n")
	for i, line := range bodyLines {
		bodyLines[i] = indentNested + line
	}
	b.WriteString(strings.Join(bodyLines, "\n"))
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
	b.WriteString(fmt.Sprintf("%s: %s  (1-%d)\n", ccLabel, ccText, runconfig.MaxConcurrency))

	ccValue, ccErr := strconv.Atoi(ccText)
	inlineConcurrencyFired := ccErr == nil && (ccValue < runconfig.MinConcurrency || ccValue > runconfig.MaxConcurrency)
	if inlineConcurrencyFired {
		b.WriteString(indentNested)
		b.WriteString(styleError.Render(fmt.Sprintf("Must be between %d and %d", runconfig.MinConcurrency, runconfig.MaxConcurrency)))
		b.WriteString("\n")
	}

	if m.errMsg != "" && !inlineConcurrencyFired {
		b.WriteString("\n")
		b.WriteString(styleError.Render(indentField + m.errMsg))
		b.WriteString("\n")
		b.WriteString(styleMuted.Render(indentField + "Adjust the request and run again."))
		b.WriteString("\n")
	}

	return b.String()
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

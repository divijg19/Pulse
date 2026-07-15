package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
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

	errs := m.fieldErrors()
	if urlErr, ok := errs["url"]; ok && urlActive {
		b.WriteString(indentNested + styleError.Render(urlErr) + "\n")
	}

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

	ccValue, ccErr := strconv.Atoi(m.concurrencyInput.Value())
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

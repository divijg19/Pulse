package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderPayloadDomain(width int) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(domainHeader("Payload", width, m.activeDomain == DomainPayload))
	b.WriteString("\n")

	m.renderHeaders(&b)

	errs := m.fieldErrors()
	if hdrErr, ok := errs["header"]; ok && m.activeDomain == DomainPayload && m.selectedHead != bodyFocus {
		b.WriteString(indentNested + styleError.Render(hdrErr) + "\n")
	}

	b.WriteString("\n")

	m.renderBody(&b)

	errs = m.fieldErrors()
	if bodyErr, ok := errs["body"]; ok && m.selectedHead == bodyFocus {
		b.WriteString(indentNested + styleError.Render(bodyErr) + "\n")
	}

	return b.String()
}

func (m Model) renderHeaders(b *strings.Builder) {
	headersActive := m.activeDomain == DomainPayload && m.selectedHead != bodyFocus
	b.WriteString(accentOrMuted(indentField+"HEADERS", headersActive))
	b.WriteString("\n")

	if len(m.headers) == 0 {
		b.WriteString(styleMuted.Render(indentNested + "No headers configured."))
		b.WriteString("\n")
		return
	}

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

func (m Model) renderBody(b *strings.Builder) {
	bodyActive := m.activeDomain == DomainPayload && m.selectedHead == bodyFocus
	bodyLen := len(m.bodyInput.Value())
	bodyLabel := "BODY"
	if bodyLen > 0 {
		bodyLabel = fmt.Sprintf("BODY (%d KB / %d KB)", bodyLen/1024, maxTUIBodyBytes/1024)
	}
	b.WriteString(indentField + accentOrMuted(bodyLabel, bodyActive))
	b.WriteString("\n")
	bodyView := m.bodyInput.View()
	bodyLines := strings.Split(bodyView, "\n")
	for i, line := range bodyLines {
		bodyLines[i] = indentNested + line
	}
	b.WriteString(strings.Join(bodyLines, "\n"))
	b.WriteString("\n")
}

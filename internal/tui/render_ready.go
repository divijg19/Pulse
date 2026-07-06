package tui

import (
	"fmt"
	"strings"

	"github.com/divijg19/Pulse/internal/runconfig"
)

func (m Model) renderReady(region Region) string {
	method := runconfig.AllowedMethods()[m.methodIndex]
	url := m.urlInput.Value()
	cc := m.concurrency()
	payloadLabel := m.payloadSummary()

	isDefaultMethod := m.methodIndex == 0
	isDefaultURL := url == "" || url == defaultURL
	isDefaultCC := cc == runconfig.DefaultConcurrency

	items := []configItem{
		{"Method", method, true},
		{"URL", url, strings.TrimSpace(url) != ""},
		{"CC", fmt.Sprintf("%d", cc), true},
		{"Payload", payloadLabel, true},
	}
	if isDefaultMethod {
		items[0].Value = styleMuted.Render(method)
	}
	if url == "" {
		items[1].Value = styleMuted.Render(sentinelEmpty)
	} else if isDefaultURL {
		items[1].Value = styleMuted.Render(url)
	}
	if isDefaultCC {
		items[2].Value = styleMuted.Render(fmt.Sprintf("%d", cc))
	}

	var b strings.Builder
	b.WriteString(styleMuted.Render("Prepare"))
	b.WriteString(gapSection)
	b.WriteString(accentOrMuted("Current Request", true))
	b.WriteString(gapSection)
	b.WriteString(renderKeyValueList(items))
	b.WriteString(gapSection)

	if m.errMsg != "" {
		b.WriteString(styleError.Render("Configuration incomplete"))
		b.WriteString("\n")
		b.WriteString(indentField + m.errMsg)
		b.WriteString("\n")
		b.WriteString(styleMuted.Render(indentField + "Press E to edit and adjust."))
	} else {
		b.WriteString("State\n")
		b.WriteString(styleMuted.Render("Ready to execute"))
	}

	return regionStyle(region).Render(b.String())
}

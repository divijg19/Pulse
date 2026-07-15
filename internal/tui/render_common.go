package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/divijg19/Pulse/internal/model"
)

func formatDuration(duration time.Duration) string {
	if duration <= 0 {
		return "0.00s"
	}
	secs := duration.Seconds()
	if secs >= 60 {
		mins := int(secs) / 60
		left := secs - float64(mins*60)
		return fmt.Sprintf("%dm %.0fs", mins, left)
	}
	return fmt.Sprintf("%.2fs", secs)
}

func renderMethod(method string) string {
	return styleMethod.Render(method)
}

func renderMetadata(label, value string) string {
	return fmt.Sprintf("%s: %s", styleMuted.Render(label), value)
}

func renderKeyValueList(items []configItem) string {
	var b strings.Builder
	for _, item := range items {
		value := item.Value
		if !item.Valid {
			value = styleError.Render(value)
		}
		b.WriteString(fmt.Sprintf(indentField+"%s: %s\n", item.Identity, value))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func renderBodyPreview(body string, maxLines int) string {
	if body == "" {
		return styleMuted.Render("No body captured.")
	}

	display := body
	trimmed := strings.TrimSpace(body)
	if json.Valid([]byte(trimmed)) {
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				display = string(formatted)
			}
		}
	}

	bodyLines := strings.Split(display, "\n")
	var b strings.Builder
	for i, line := range bodyLines {
		if maxLines > 0 && i >= maxLines {
			b.WriteString(styleMuted.Render("... (truncated)"))
			break
		}
		b.WriteString(line)
		if i < len(bodyLines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
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

func identityCell(label string) string {
	return styleMuted.Render(label)
}

// sectionLine renders a section header with em-dash rules around the label.
// When active is true, the accent style is used; otherwise muted.
func sectionLine(label string, active bool) string {
	if active {
		return styleAccent.Render("── " + label + " ──")
	}
	return styleSectionLine.Render("── " + label + " ──")
}

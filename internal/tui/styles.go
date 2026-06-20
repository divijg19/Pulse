package tui

import "github.com/charmbracelet/lipgloss"

const (
	colorBg      = "#09090b"
	colorText    = "#e4e4e7"
	colorMuted   = "#a1a1aa"
	colorDark    = "#27272a"
	colorAccent  = "#38bdf8"
	colorSuccess = "#34d399"
	colorWarning = "#fbbf24"
	colorError   = "#f87171"
)

var (
	StyleBase      = lipgloss.NewStyle().Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorText))
	StyleTopBar    = StyleBase.Copy().Bold(true)
	StyleStatusBar = StyleBase.Copy().Background(lipgloss.Color(colorDark))
	StyleStatusMode = lipgloss.NewStyle().Background(lipgloss.Color(colorAccent)).Foreground(lipgloss.Color(colorBg)).Bold(true)
	StyleSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
)

func statusColor(status int) string {
	if status == 0 {
		return colorError
	}
	if status >= 200 && status < 300 {
		return colorSuccess
	}
	if status >= 300 && status < 400 {
		return colorWarning
	}
	if status >= 400 {
		return colorError
	}
	return colorMuted
}

func rowStyle(selected bool) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	if selected {
		style = style.Foreground(lipgloss.Color(colorAccent)).Background(lipgloss.Color(colorDark))
	}
	return style
}

func errorRowStyle(selected bool) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
	if selected {
		style = style.Background(lipgloss.Color(colorDark)).Bold(true)
	}
	return style
}

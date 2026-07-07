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
	styleBase   = lipgloss.NewStyle().Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorText))
	styleTopBar = styleBase.Bold(true)
	styleRibbon = styleBase
	styleMuted  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleAccent = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	styleError  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))

	styleWorkspaceBadge = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorBg)).
				Background(lipgloss.Color(colorAccent)).
				Bold(true).
				Padding(0, 1)

	stylePrimaryAction = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent)).
				Background(lipgloss.Color(colorDark)).
				Bold(true)

	styleStatusCell = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Bold(true)

	styleSectionLine = styleMuted

	styleCompareMarked = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning)).Bold(true)
	styleCompareActive = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
	styleDiffAdded     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess))
	styleDiffRemoved   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))

	styleMethod = styleBase.Bold(true)

	styleDomainActive = styleAccent

	styleDomainInactive = styleMuted
)

func regionStyle(region Region) lipgloss.Style {
	return styleBase.Width(region.Width).Height(region.Height)
}

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

var (
	rowStyleSelected   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Background(lipgloss.Color(colorDark))
	rowStyleUnselected = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))

	errorRowStyleSelected   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Background(lipgloss.Color(colorDark)).Bold(true)
	errorRowStyleUnselected = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
)

func rowStyle(selected bool) lipgloss.Style {
	if selected {
		return rowStyleSelected
	}
	return rowStyleUnselected
}

func errorRowStyle(selected bool) lipgloss.Style {
	if selected {
		return errorRowStyleSelected
	}
	return errorRowStyleUnselected
}

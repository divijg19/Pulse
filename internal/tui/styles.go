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
	styleBase      = lipgloss.NewStyle().Background(lipgloss.Color(colorBg)).Foreground(lipgloss.Color(colorText))
	styleTopBar    = styleBase.Copy().Bold(true)
	styleRibbon    = styleBase.Copy().Background(lipgloss.Color(colorDark))
	styleSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleMuted     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleAccent    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent))
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))

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

	styleSectionLine = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMuted))

	styleHeading = lipgloss.NewStyle().
			Bold(true)

	styleDomainActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent))

	styleDomainInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMuted))
)

func regionStyle(region Region) lipgloss.Style {
	return styleBase.Copy().Width(region.Width).Height(region.Height)
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

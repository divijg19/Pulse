package tui

import "github.com/charmbracelet/lipgloss"

const (
	colorBg      = "#09090b"
	colorPanel   = "#111113"
	colorBorder  = "#27272a"
	colorText    = "#d4d4d8"
	colorMuted   = "#71717a"
	colorCyan    = "#22d3ee"
	colorGreen   = "#34d399"
	colorAmber   = "#fbbf24"
	colorRose    = "#fb7185"
	colorFuchsia = "#e879f9"
)

var (
	appStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText)).
			Background(lipgloss.Color(colorBg)).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f4f4f5")).
			Bold(true)

	statusDotStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	monoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Bold(true)

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Bold(true)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Background(lipgloss.Color(colorPanel)).
			Padding(1, 2)

	metricStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color(colorBorder)).
			Foreground(lipgloss.Color(colorText)).
			PaddingLeft(1)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorCyan)).
				Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Align(lipgloss.Center, lipgloss.Center)

	errorTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorRose))
)

func controlStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 1)
	if focused {
		style = style.BorderForeground(lipgloss.Color(colorCyan)).Foreground(lipgloss.Color(colorCyan))
	}
	return style
}

func inputStyle(focused bool) lipgloss.Style {
	style := controlStyle(focused)
	if !focused {
		style = style.Foreground(lipgloss.Color("#e4e4e7"))
	}
	return style
}

func bodyStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(0, 1)
	if focused {
		style = style.BorderForeground(lipgloss.Color(colorCyan))
	}
	return style
}

func runButtonStyle(running bool) lipgloss.Style {
	if running {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorRose)).
			Foreground(lipgloss.Color(colorRose)).
			Bold(true).
			Align(lipgloss.Center)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#f4f4f5")).
		Foreground(lipgloss.Color("#09090b")).
		Background(lipgloss.Color("#f4f4f5")).
		Bold(true).
		Align(lipgloss.Center)
}

func tabStyle(active bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Padding(0, 3).
		Foreground(lipgloss.Color(colorMuted))
	if active {
		style = style.Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#27272a")).Bold(true)
	}
	return style
}

func rowStyle(selected bool) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	if selected {
		style = style.Foreground(lipgloss.Color(colorCyan)).Background(lipgloss.Color("#18181b"))
	}
	return style
}

func errorRowStyle(selected bool) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorRose))
	if selected {
		style = style.Background(lipgloss.Color("#2a1218")).Bold(true)
	}
	return style
}

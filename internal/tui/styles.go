package tui

import "github.com/charmbracelet/lipgloss"

const (
	colorBg         = "#09090b"
	colorPanel      = "#111113"
	colorElevated   = "#1c1c1e"
	colorBorder     = "#3f3f46"
	colorText       = "#d4d4d8"
	colorMuted      = "#71717a"
	colorCyan       = "#22d3ee"
	colorCyanStrong = "#06b6d4"
	colorGreen      = "#34d399"
	colorAmber      = "#fbbf24"
	colorRose       = "#fb7185"
	colorFuchsia    = "#e879f9"
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

	statusDotGlowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorCyanStrong)).
				Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Bold(true)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Background(lipgloss.Color(colorPanel)).
			Padding(1, 2)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorCyan)).
				Bold(true)

	separatorBorder = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	metricValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorText)).
				Bold(true)

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

func pillStyle(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colorBorder)).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Padding(0, 3).
		Foreground(lipgloss.Color(colorMuted))
}

func rowStyle(selected bool) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	if selected {
		style = style.Foreground(lipgloss.Color(colorCyan)).Background(lipgloss.Color(colorElevated))
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

func methodColor(name string) string {
	switch name {
	case "GET":
		return colorCyan
	case "POST":
		return colorGreen
	case "PUT":
		return colorAmber
	case "DELETE":
		return colorRose
	case "PATCH":
		return colorFuchsia
	default:
		return colorMuted
	}
}

func methodStyle(name string, focused bool) lipgloss.Style {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(0, 1)
	if focused {
		s = s.BorderForeground(lipgloss.Color(colorCyan)).Foreground(lipgloss.Color(methodColor(name)))
	} else {
		s = s.Foreground(lipgloss.Color(methodColor(name)))
	}
	return s
}

func statusColor(status int) string {
	if status == 0 {
		return colorRose
	}
	if status >= 200 && status < 300 {
		return colorGreen
	}
	if status >= 300 && status < 400 {
		return colorCyan
	}
	if status >= 400 {
		return colorRose
	}
	return colorMuted
}

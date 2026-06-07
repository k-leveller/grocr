package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen  = lipgloss.Color("#00cc66")
	ColorRed    = lipgloss.Color("#ff4444")
	ColorYellow = lipgloss.Color("#ffcc00")
	ColorCyan   = lipgloss.Color("#00cccc")
	ColorDim    = lipgloss.Color("#666666")
	ColorWhite  = lipgloss.Color("#ffffff")

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	StyleModeAdd = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)

	StyleModeConsume = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorRed)

	StyleHint = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorYellow)

	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorCyan)

	StyleBold = lipgloss.NewStyle().
			Bold(true)

	StyleLabel = lipgloss.NewStyle().
			Bold(true).
			Width(12)

	StyleSeparator = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleRequired = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleLogEntry = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleHelpOverlay = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorCyan).
				Padding(1, 2)
)

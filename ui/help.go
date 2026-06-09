package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/k-leveller/grocr/locale"
)

func RenderHelp(width, height int) string {
	content := StyleBold.Render(locale.Active.HelpTitle) + "\n\n" +
		locale.Active.HelpBody +
		StyleBold.Render(locale.Active.HelpUnknownBarcodeTitle) + "\n\n" +
		locale.Active.HelpUnknownBarcodeBody

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

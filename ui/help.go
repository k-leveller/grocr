package ui

import "github.com/charmbracelet/lipgloss"

func RenderHelp(width, height int) string {
	content := StyleBold.Render("Keyboard Shortcuts") + "\n\n" +
		"  q       Quit\n" +
		"  m       Toggle Add/Consume mode\n" +
		"  /       Search products by name\n" +
		"  e       Edit product name\n" +
		"  ?       Toggle this help\n" +
		"  Tab     Next field\n" +
		"  S-Tab   Previous field\n" +
		"  Enter   Advance / Submit\n" +
		"  Esc     Cancel current scan\n" +
		"  Ctrl+C  Quit\n"

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

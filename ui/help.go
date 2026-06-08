package ui

import "github.com/charmbracelet/lipgloss"

func RenderHelp(width, height int) string {
	content := StyleBold.Render("Keyboard Shortcuts") + "\n\n" +
		"  q       Quit\n" +
		"  m       Toggle Add/Consume mode\n" +
		"  /       Search products by name\n" +
		"  Ctrl+N  New product (no barcode)\n" +
		"  Ctrl+E  Export stock to CSV\n" +
		"  e       Edit product name\n" +
		"  p       Price history (lookup view)\n" +
		"  ?       Toggle this help\n" +
		"  ↑/↓     Cycle UPC scan history (idle input)\n" +
		"  j/k     Navigate expiring-soon panel\n" +
		"  Tab/↓   Next field (form)\n" +
		"  S-Tab/↑ Previous field (form)\n" +
		"  Enter   Submit form\n" +
		"  Esc     Cancel current scan\n" +
		"  Ctrl+C  Quit\n"

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

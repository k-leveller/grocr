package ui

import "github.com/charmbracelet/lipgloss"

func RenderHelp(width, height int) string {
	content := StyleBold.Render("Keyboard Shortcuts") + "\n\n" +
		"  q       Quit\n" +
		"  m       Toggle Add/Consume mode\n" +
		"  /       Search products by name\n" +
		"  Ctrl+N  New product (no barcode)\n" +
		"  Ctrl+E  Export stock to CSV\n" +
		"  t       Today's meal plan (compact)\n" +
		"  P       Meal plan (next 7 days, scrollable)\n" +
		"  r       Recipe list (fulfillability)\n" +
		"  e       Edit product name\n" +
		"  p       Price history (lookup view)\n" +
		"  t       Transfer stock between locations (lookup view)\n" +
		"  ?       Toggle this help\n" +
		"  ↑/↓     Cycle UPC scan history (idle input)\n" +
		"  j/k     Navigate expiring-soon panel\n" +
		"  d       Mark as spoiled (expiring-soon panel)\n" +
		"  Tab/↓   Next field (form)\n" +
		"  S-Tab/↑ Previous field (form)\n" +
		"  Enter   Submit form\n" +
		"  Esc     Cancel current scan\n" +
		"  Ctrl+C  Quit\n\n" +
		StyleBold.Render("Unknown Barcode") + "\n\n" +
		"  C/Enter  Create a new product for this barcode\n" +
		"  L        Link barcode to an existing product (rebrand/repackage)\n"

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

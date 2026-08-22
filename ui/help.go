package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/k-leveller/grocr/keybind"
	"github.com/k-leveller/grocr/locale"
)

// helpOrder is the order rebindable actions appear in the help overlay.
var helpOrder = []keybind.Action{
	keybind.Quit,
	keybind.Mode,
	keybind.Search,
	keybind.NewProduct,
	keybind.Export,
	keybind.MealPlanToday,
	keybind.MealPlan,
	keybind.Recipes,
	keybind.EditName,
	keybind.Notes,
	keybind.PriceHistory,
	keybind.Transfer,
	keybind.Help,
	keybind.Up,
	keybind.Down,
	keybind.Left,
	keybind.Right,
	keybind.Consume,
	keybind.Spoil,
	keybind.Refresh,
}

type helpRow struct{ key, desc string }

// helpRows builds the key/description lines for every bound action, followed by
// the fixed shortcuts. All rows share one key column, sized to the widest key.
func helpRows() string {
	rows := make([]helpRow, 0, len(helpOrder))
	for _, a := range helpOrder {
		desc, ok := locale.Active.HelpDescs[a]
		if !ok {
			continue
		}
		rows = append(rows, helpRow{keyWithArrow(a), desc})
	}
	bound := len(rows)
	rows = append(rows, staticHelpRows()...)

	width := 0
	for _, r := range rows {
		if n := len([]rune(r.key)); n > width {
			width = n
		}
	}

	var b strings.Builder
	for i, r := range rows {
		if i == bound {
			b.WriteString("\n")
		}
		b.WriteString("  " + padRight(r.key, width+2) + r.desc + "\n")
	}
	return b.String()
}

// staticHelpRows parses the locale's tab-separated fixed-shortcut lines.
func staticHelpRows() []helpRow {
	var rows []helpRow
	for _, line := range strings.Split(locale.Active.HelpStaticRows, "\n") {
		key, desc, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		rows = append(rows, helpRow{key, desc})
	}
	return rows
}

func RenderHelp(width, height int) string {
	create, link := padPair(keybind.Sym(keybind.Create), keybind.Sym(keybind.Link))

	content := StyleBold.Render(locale.Active.HelpTitle) + "\n\n" +
		helpRows() + "\n" +
		StyleBold.Render(locale.Active.HelpUnknownBarcodeTitle) + "\n\n" +
		fmt.Sprintf(locale.Active.HelpUnknownBarcodeBody, create, link)

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

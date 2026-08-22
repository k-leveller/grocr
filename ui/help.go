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

// helpRows builds the key/description lines for every bound action, aligned to
// the widest key column.
func helpRows() string {
	type row struct{ key, desc string }

	rows := make([]row, 0, len(helpOrder))
	width := 0
	for _, a := range helpOrder {
		desc, ok := locale.Active.HelpDescs[a]
		if !ok {
			continue
		}
		k := keyWithArrow(a)
		if n := len([]rune(k)); n > width {
			width = n
		}
		rows = append(rows, row{k, desc})
	}

	var b strings.Builder
	for _, r := range rows {
		b.WriteString("  " + padRight(r.key, width+2) + r.desc + "\n")
	}
	return b.String()
}

func RenderHelp(width, height int) string {
	create, link := padPair(keybind.Sym(keybind.Create), keybind.Sym(keybind.Link))

	content := StyleBold.Render(locale.Active.HelpTitle) + "\n\n" +
		helpRows() +
		locale.Active.HelpStaticBody +
		StyleBold.Render(locale.Active.HelpUnknownBarcodeTitle) + "\n\n" +
		fmt.Sprintf(locale.Active.HelpUnknownBarcodeBody, create, link)

	overlay := StyleHelpOverlay.Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

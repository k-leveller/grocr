package ui

import (
	"fmt"
	"strings"

	"github.com/k-leveller/grocr/keybind"
	"github.com/k-leveller/grocr/locale"
)

// This file resolves the locale's hint format strings against the active
// keybinds, so every hint shown in the UI reflects the user's own bindings.

func sym(a keybind.Action) string { return keybind.Sym(a) }

// arrowAlias lists the arrow key that keeps working alongside a bound key.
var arrowAlias = map[keybind.Action]string{
	keybind.Up:    "↑",
	keybind.Down:  "↓",
	keybind.Left:  "←",
	keybind.Right: "→",
}

// keyWithArrow renders a navigation binding together with its arrow alias,
// e.g. "→/l", collapsing to just the arrow when the two are the same.
func keyWithArrow(a keybind.Action) string {
	s := sym(a)
	if arrow, ok := arrowAlias[a]; ok && s != arrow {
		return arrow + "/" + s
	}
	return s
}

// HeaderHints is the key hint line shown on the right of the header bar.
func HeaderHints() string {
	return fmt.Sprintf(locale.Active.HeaderHints,
		sym(keybind.Quit), sym(keybind.Mode), sym(keybind.Search),
		sym(keybind.NewProduct), sym(keybind.Help))
}

// HintLookupView is the input-line hint for the product overview.
func HintLookupView() string {
	return fmt.Sprintf(locale.Active.HintLookupView,
		sym(keybind.Notes), sym(keybind.PriceHistory), sym(keybind.Transfer))
}

// HintPriceHistory is the input-line hint for the price history view.
func HintPriceHistory() string {
	return fmt.Sprintf(locale.Active.HintPriceHistory,
		sym(keybind.Down), sym(keybind.Up), sym(keybind.PriceHistory))
}

// HintMealPlan is the input-line hint for the 7-day meal plan view.
func HintMealPlan() string {
	return fmt.Sprintf(locale.Active.HintMealPlan,
		sym(keybind.Down), sym(keybind.Up), sym(keybind.Refresh), sym(keybind.Quit))
}

// HintTodayMealPlan is the input-line hint for today's meal plan.
func HintTodayMealPlan() string {
	return fmt.Sprintf(locale.Active.HintTodayMealPlan,
		sym(keybind.Refresh), sym(keybind.Quit))
}

// HintRecipeList is the input-line hint for the recipe list.
func HintRecipeList() string {
	return fmt.Sprintf(locale.Active.HintRecipeList,
		sym(keybind.Down), sym(keybind.Up), sym(keybind.Search),
		keyWithArrow(keybind.Right), sym(keybind.Refresh), sym(keybind.Quit))
}

// HintRecipeDetail is the input-line hint for a single recipe.
func HintRecipeDetail() string {
	return fmt.Sprintf(locale.Active.HintRecipeDetail,
		sym(keybind.Down), sym(keybind.Up), keyWithArrow(keybind.Left), sym(keybind.Quit))
}

// HintYesNo is the input-line hint for the shopping-list prompt.
func HintYesNo() string {
	return fmt.Sprintf(locale.Active.HintYesNo, sym(keybind.Yes))
}

// HintUnknownBarcode is the input-line hint for the unknown-barcode prompt.
func HintUnknownBarcode() string {
	return fmt.Sprintf(locale.Active.HintUnknownBarcode,
		sym(keybind.Create), sym(keybind.Link))
}

// HintExpiringDetail is the input-line hint for the expiring-item detail panel.
func HintExpiringDetail() string {
	return fmt.Sprintf(locale.Active.HintExpiringDetail,
		keyWithArrow(keybind.Left), sym(keybind.Down), sym(keybind.Up))
}

// SearchNavHint is the navigation hint shown below the search results.
func SearchNavHint() string {
	return fmt.Sprintf(locale.Active.SearchNavHint, sym(keybind.NewProduct))
}

// ShoppingListPrompt asks whether to add the just-emptied product to the list.
func ShoppingListPrompt() string {
	return fmt.Sprintf(locale.Active.ShoppingListPrompt, sym(keybind.Yes))
}

// UnknownBarcodeActions returns the two action lines of the unknown-barcode
// prompt. The key symbols are padded to a common width so the descriptions stay
// column-aligned whatever the bindings are.
func UnknownBarcodeActions() (create, link string) {
	c, l := padPair(sym(keybind.Create), sym(keybind.Link))
	return fmt.Sprintf(locale.Active.UnknownBarcodeCreate, c),
		fmt.Sprintf(locale.Active.UnknownBarcodeLink, l)
}

// padPair right-pads both strings to their common width.
func padPair(a, b string) (string, string) {
	w := max(len([]rune(a)), len([]rune(b)))
	return padRight(a, w), padRight(b, w)
}

func padRight(s string, w int) string {
	if n := w - len([]rune(s)); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

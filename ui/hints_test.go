package ui

import (
	"strings"
	"testing"

	"github.com/k-leveller/grocr/keybind"
	"github.com/k-leveller/grocr/locale"
)

// hintFuncs is every hint that interpolates keybinds.
var hintFuncs = map[string]func() string{
	"HeaderHints":        HeaderHints,
	"HintLookupView":     HintLookupView,
	"HintPriceHistory":   HintPriceHistory,
	"HintMealPlan":       HintMealPlan,
	"HintTodayMealPlan":  HintTodayMealPlan,
	"HintRecipeList":     HintRecipeList,
	"HintRecipeDetail":   HintRecipeDetail,
	"HintYesNo":          HintYesNo,
	"HintUnknownBarcode": HintUnknownBarcode,
	"HintExpiringDetail": HintExpiringDetail,
	"SearchNavHint":      SearchNavHint,
	"ShoppingListPrompt": ShoppingListPrompt,
}

// TestHintsFormatCleanly guards against a hint format string whose verb count
// drifts from the arguments ui/hints.go passes it.
func TestHintsFormatCleanly(t *testing.T) {
	for name, fn := range hintFuncs {
		got := fn()
		if strings.Contains(got, "%!") || strings.Contains(got, "%s") {
			t.Errorf("%s produced a malformed string: %q", name, got)
		}
	}

	create, link := UnknownBarcodeActions()
	for _, s := range []string{create, link} {
		if strings.Contains(s, "%!") || strings.Contains(s, "%s") {
			t.Errorf("UnknownBarcodeActions produced a malformed string: %q", s)
		}
	}
}

func TestHintsUseActiveKeybinds(t *testing.T) {
	orig := keybind.Active
	t.Cleanup(func() { keybind.Active = orig })
	keybind.Active = keybind.Parse("quit = ctrl+q\nmode = M\nnotes = N\n")

	if got := HeaderHints(); !strings.Contains(got, "Ctrl+Q:quit") || !strings.Contains(got, "M:mode") {
		t.Errorf("HeaderHints = %q, want the rebound quit and mode keys", got)
	}
	if got := HintLookupView(); !strings.Contains(got, "N = notes") {
		t.Errorf("HintLookupView = %q, want the rebound notes key", got)
	}
	if got := SearchNavHint(); !strings.Contains(got, "n new") {
		t.Errorf("SearchNavHint = %q, want the default new-product key", got)
	}
}

func TestUnknownBarcodeActionsAlign(t *testing.T) {
	orig := keybind.Active
	t.Cleanup(func() { keybind.Active = orig })
	keybind.Active = keybind.Parse("create = ctrl+n\nlink = l\n")

	create, link := UnknownBarcodeActions()
	if !strings.Contains(create, "Ctrl+N") || !strings.Contains(link, "l") {
		t.Fatalf("bindings not interpolated: %q / %q", create, link)
	}
	if strings.Index(create, "]") != strings.Index(link, "]") {
		t.Errorf("key columns are not aligned:\n%q\n%q", create, link)
	}
}

func TestHelpRowsCoverEveryListedAction(t *testing.T) {
	body := helpRows()
	for _, a := range helpOrder {
		desc, ok := locale.Active.HelpDescs[a]
		if !ok {
			t.Errorf("action %q is listed in helpOrder but has no description", a)
			continue
		}
		if !strings.Contains(body, desc) {
			t.Errorf("help overlay is missing the row for %q", a)
		}
	}
	for a := range locale.Active.HelpDescs {
		if !containsAction(helpOrder, a) {
			t.Errorf("action %q has a description but is not listed in helpOrder", a)
		}
	}
}

func TestHelpRowsShowActiveKeys(t *testing.T) {
	orig := keybind.Active
	t.Cleanup(func() { keybind.Active = orig })
	keybind.Active = keybind.Parse("quit = ctrl+q\nup = w\n")

	body := helpRows()
	if !strings.Contains(body, "Ctrl+Q") {
		t.Errorf("help rows do not show the rebound quit key:\n%s", body)
	}
	if !strings.Contains(body, "↑/w") {
		t.Errorf("help rows do not show the arrow alias for up:\n%s", body)
	}
}

func TestKeyWithArrowDropsRedundantArrow(t *testing.T) {
	orig := keybind.Active
	t.Cleanup(func() { keybind.Active = orig })
	keybind.Active = keybind.Parse("up = up\n")

	if got := keyWithArrow(keybind.Up); got != "↑" {
		t.Errorf("keyWithArrow(Up) = %q, want ↑", got)
	}
}

func containsAction(list []keybind.Action, a keybind.Action) bool {
	for _, x := range list {
		if x == a {
			return true
		}
	}
	return false
}
